// /home/krylon/go/src/github.com/blicero/hertz/web/web.go
// -*- mode: go; coding: utf-8; -*-
// Created on 03. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-09 11:36:42 krylon>

package web

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/blicero/hertz/common"
	"github.com/blicero/hertz/database"
	"github.com/blicero/hertz/logdomain"
	"github.com/blicero/hertz/model"
	"github.com/gorilla/mux"
)

const (
	cacheControl = "max-age=3600, public"
	noCache      = "no-store, max-age=0"
	tmplFolder   = "assets/templates"
	poolSize     = 4
)

//go:embed assets
var assets embed.FS

// Server provides a web-based UI
type Server struct {
	addr      string
	log       *log.Logger
	lock      sync.RWMutex // nolint: unused
	active    atomic.Bool
	router    *mux.Router
	tmpl      *template.Template
	web       http.Server
	mimeTypes map[string]string
	pool      *database.Pool
}

// Create returns a new web Server.
func Create(addr string) (*Server, error) {
	var (
		err error
		msg string
		srv = &Server{
			addr: addr,
			mimeTypes: map[string]string{
				".css":  "text/css",
				".map":  "application/json",
				".js":   "text/javascript",
				".png":  "image/png",
				".jpg":  "image/jpeg",
				".jpeg": "image/jpeg",
				".webp": "image/webp",
				".gif":  "image/gif",
				".json": "application/json",
				".html": "text/html",
			},
		}
	)

	if srv.log, err = common.GetLogger(logdomain.Web); err != nil {
		return nil, err
	} else if srv.pool, err = database.NewPool(poolSize); err != nil {
		srv.log.Printf("[CRITICAL] Failed to open Database connection pool: %s\n",
			err.Error())
		return nil, err
	}

	var templates []fs.DirEntry
	var tmplRe = regexp.MustCompile("[.]tmpl$")

	if templates, err = assets.ReadDir(tmplFolder); err != nil {
		srv.log.Printf("[ERROR] Cannot read embedded templates: %s\n",
			err.Error())
		return nil, err
	}

	srv.tmpl = template.New("").Funcs(funcmap)
	for _, entry := range templates {
		var (
			content []byte
			path    = filepath.Join(tmplFolder, entry.Name())
		)

		if !tmplRe.MatchString(entry.Name()) {
			continue
		} else if content, err = assets.ReadFile(path); err != nil {
			msg = fmt.Sprintf("Cannot read embedded file %s: %s",
				path,
				err.Error())
			srv.log.Printf("[CRITICAL] %s\n", msg)
			return nil, errors.New(msg)
		} else if srv.tmpl, err = srv.tmpl.Parse(string(content)); err != nil {
			msg = fmt.Sprintf("Could not parse template %s: %s",
				entry.Name(),
				err.Error())
			srv.log.Println("[CRITICAL] " + msg)
			return nil, errors.New(msg)
		} else if common.Debug {
			srv.log.Printf("[TRACE] Template \"%s\" was parsed successfully.\n",
				entry.Name())
		}
	}

	srv.router = mux.NewRouter()
	srv.web.Addr = addr
	srv.web.ErrorLog = srv.log
	srv.web.Handler = srv.router

	// Register URL handlers
	srv.router.NotFoundHandler = http.HandlerFunc(srv.handleNotFound)
	srv.router.HandleFunc("/favicon.ico", srv.handleFavIco)
	srv.router.HandleFunc("/static/{file}", srv.handleStaticFile)
	srv.router.HandleFunc("/{index:(?i:index|main|start)$}", srv.handleMain)

	// AJAX Handlers
	srv.router.HandleFunc(
		"/ajax/beacon",
		srv.handleBeacon)

	// Web service endpoints
	srv.router.HandleFunc("/ws/get_timestamp/{name:(?:\\w+)$}",
		srv.handleClientGetTimestamp)

	srv.router.HandleFunc("/ws/submit_data/{name:(?:\\w+)$}",
		srv.handleClientData)

	return srv, nil
} // func Create(addr string) (*Server, error)

// IsActive returns the Server's active flag.
func (srv *Server) IsActive() bool {
	return srv.active.Load()
} // func (srv *Server) IsActive() bool

// Stop clears the Server's active flag.
func (srv *Server) Stop() {
	srv.active.Store(false)
	srv.web.Shutdown(context.Background())
	srv.pool.Close()
} // func (srv *Server) Stop()

// Run executes the Server's loop, waiting for new connections and starting
// goroutines to handle them.
func (srv *Server) Run() {
	var err error

	defer srv.log.Println("[INFO] Web server is shutting down")

	srv.active.Store(true)
	defer srv.active.Store(false)

	srv.log.Printf("[INFO] Web frontend is going online at %s\n", srv.addr)
	http.Handle("/", srv.router)

	if err = srv.web.ListenAndServe(); err != nil {
		if err.Error() != "http: Server closed" {
			srv.log.Printf("[ERROR] ListenAndServe returned an error: %s\n",
				err.Error())
		} else {
			srv.log.Println("[INFO] HTTP Server has shut down.")
		}
	}
} // func (srv *Server) Run()

//////////////////////////////////////////////////////////////////////////////
/// Handle requests //////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////

func (srv *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handling request for %s\n", r.RequestURI)
	srv.log.Printf("[ERROR] 404 - %s\n", r.RequestURI)

	srv.sendErrorMessage(
		w,
		fmt.Sprintf(
			"No Handler was found for %s",
			r.RequestURI))
} // func (srv *Server) handleNotFound(w http.ResponseWriter, r *http.Request)

func (srv *Server) handleMain(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handling request for %s\n", r.RequestURI)
	const tmplName = "main"

	var (
		err  error
		msg  string
		tmpl *template.Template
		data = tmplDataIndex{
			tmplDataBase: tmplDataBase{
				Title: "Main",
				Debug: common.Debug,
				URL:   r.URL.String(),
			},
		}
	)

	if tmpl = srv.tmpl.Lookup(tmplName); tmpl == nil {
		msg = fmt.Sprintf("Could not find template %q", tmplName)
		srv.log.Println("[CRITICAL] " + msg)
		srv.sendErrorMessage(w, msg)
		return
	}

	w.Header().Set("Cache-Control", noCache)
	if err = tmpl.Execute(w, &data); err != nil {
		msg = fmt.Sprintf("Error rendering template %q: %s",
			tmplName,
			err.Error())
		srv.sendErrorMessage(w, msg)
	}
} // func (srv *Server) handleMain(w http.ResponseWriter, r *http.Request)

//////////////////////////////////////////////////////////////////////////////
/// Handle AJAX //////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////

func (srv *Server) handleBeacon(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		buf  []byte
		data = ajaxBeaconData{
			ajaxData: ajaxData{
				Status:    true,
				Timestamp: time.Now(),
				Message:   common.AppName + " " + common.Version,
			},
			Hostname: hostname(),
		}
	)

	if buf, err = json.Marshal(&data); err != nil {
		var msg = fmt.Sprintf("Failed to serialize payload for AJAX response: %s",
			err.Error())
		srv.log.Printf("[CANTHAPPEN] %s\n", msg)
		buf = errJSON(msg)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", noCache)
	w.WriteHeader(200)
	w.Write(buf) // nolint: errcheck,gosec
} // func (srv *Server) handleBeacon(w http.ResponseWriter, r *http.Request)

//////////////////////////////////////////////////////////////////////////////
/// Handle Clients ///////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////

func (srv *Server) handleClientGetTimestamp(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handling request for %s\n", r.RequestURI)

	var (
		err       error
		vars      map[string]string
		name, msg string
		host      *model.Host
		db        *database.Database

		buf  []byte
		data = model.SrvResponse{
			Timestamp: time.Now(),
		}
	)

	vars = mux.Vars(r)
	name = vars["name"]

	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if host, err = db.HostGetByName(name); err != nil {
		msg = fmt.Sprintf("Failed to lookup Client %s: %s",
			name,
			err.Error())
		srv.log.Printf("[ERROR] %s\n", msg)
		data.Message = msg
	} else if host == nil {
		data.Payload = "0"
		data.Status = true
	} else {
		data.Payload = strconv.FormatInt(host.LastContact.Unix(), 10)
		data.Status = true
	}

	if buf, err = json.Marshal(&data); err != nil {
		msg = fmt.Sprintf("Failed to serialize Response: %s\n",
			err.Error())
		buf = errJSON(msg)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", noCache)
	w.WriteHeader(200)
	w.Write(buf) // nolint: errcheck,gosec
} // func (srv *Server) handleClientGetTimestamp(w http.ResponseWriter, r *http.Request)

func (srv *Server) handleClientData(w http.ResponseWriter, r *http.Request) {
	srv.log.Printf("[TRACE] Handling request for %s\n", r.RequestURI)

	var (
		err       error
		db        *database.Database
		vars      map[string]string
		name, msg string
		body      []byte
		buf       []byte
		records   []*model.Record
		res       = model.SrvResponse{
			Timestamp: time.Now(),
		}
	)

	vars = mux.Vars(r)
	name = vars["name"]
	records = make([]*model.Record, 0)

	if body, err = io.ReadAll(r.Body); err != nil {
		msg = fmt.Sprintf("failed to read request body: %s",
			err.Error())
		srv.log.Printf("[CANTHAPPEN] %s\n",
			msg)
		res.Message = msg
		goto SEND
	} else if err = json.Unmarshal(body, &records); err != nil {
		msg = fmt.Sprintf("failed to parse request body: %s\n%s\n\n",
			err.Error(),
			body)
		srv.log.Printf("[ERROR] %s\n", msg)
		res.Message = msg
		goto SEND
	}

	srv.log.Printf("[DEBUG] Received %d Records from %s\n",
		len(records),
		name)

	db = srv.pool.Get()
	defer srv.pool.Put(db)

	if err = db.RecordAddBulk(name, records); err != nil {
		msg = fmt.Sprintf("Failed to store %d records from %s: %s",
			len(records),
			name,
			err.Error())
		srv.log.Printf("[ERROR] %s\n", msg)
		res.Message = msg
	}

	res.Status = true
	res.Message = fmt.Sprintf("Successfully stored %d records", len(records))
	res.Payload = strconv.FormatInt(
		records[len(records)-1].Timestamp.Unix(),
		10)

SEND:
	if buf, err = json.Marshal(&res); err != nil {
		msg = fmt.Sprintf("Failed to serialize response: %s",
			err.Error())
		srv.log.Printf("[ERROR] %s\n",
			msg)
		buf = errJSON(msg)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", noCache)
	w.WriteHeader(200)
	w.Write(buf) // nolint: errcheck,gosec
} // func (srv *Server) handleClientData(w http.ResponseWriter, r *http.Request)

//////////////////////////////////////////////////////////////////////////////
/// Handle static assets /////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////

func (srv *Server) handleFavIco(w http.ResponseWriter, request *http.Request) {
	srv.log.Printf("[TRACE] Handle request for %s\n",
		request.URL.EscapedPath())

	const (
		filename = "assets/static/favicon.ico"
		mimeType = "image/vnd.microsoft.icon"
	)

	w.Header().Set("Content-Type", mimeType)

	if !common.Debug {
		w.Header().Set("Cache-Control", cacheControl)
	} else {
		w.Header().Set("Cache-Control", noCache)
	}

	var (
		err error
		fh  fs.File
	)

	if fh, err = assets.Open(filename); err != nil {
		msg := fmt.Sprintf("ERROR - cannot find file %s", filename)
		srv.sendErrorMessage(w, msg)
	} else {
		defer fh.Close() // nolint: errcheck
		w.WriteHeader(200)
		io.Copy(w, fh) // nolint: errcheck
	}
} // func (srv *Server) handleFavIco(w http.ResponseWriter, request *http.Request)

func (srv *Server) handleStaticFile(w http.ResponseWriter, request *http.Request) {
	// srv.log.Printf("[TRACE] Handle request for %s\n",
	// 	request.URL.EscapedPath())

	// Since we controll what static files the server has available, we
	// can easily map MIME type to slice. Soon.

	vars := mux.Vars(request)
	filename := vars["file"]
	path := filepath.Join("assets", "static", filename)

	var mimeType string

	// srv.log.Printf("[TRACE] Delivering static file %s to client %s\n",
	// 	filename,
	// 	request.RemoteAddr)

	var match []string

	if match = common.SuffixPattern.FindStringSubmatch(filename); match == nil {
		mimeType = "text/plain"
	} else if mime, ok := srv.mimeTypes[match[1]]; ok {
		mimeType = mime
	} else {
		srv.log.Printf("[ERROR] Did not find MIME type for %s\n", filename)
	}

	w.Header().Set("Content-Type", mimeType)

	if common.Debug {
		w.Header().Set("Cache-Control", noCache)
	} else {
		w.Header().Set("Cache-Control", cacheControl)
	}

	var (
		err error
		fh  fs.File
	)

	if fh, err = assets.Open(path); err != nil {
		msg := fmt.Sprintf("ERROR - cannot find file %s", path)
		srv.sendErrorMessage(w, msg)
	} else {
		defer fh.Close() // nolint: errcheck
		w.WriteHeader(200)
		io.Copy(w, fh) // nolint: errcheck
	}
} // func (srv *Server) handleStaticFile(w http.ResponseWriter, request *http.Request)

func (srv *Server) sendErrorMessage(w http.ResponseWriter, msg string) {
	html := `
<!DOCTYPE html>
<html>
  <head>
    <title>Internal Error</title>
  </head>
  <body>
    <h1>Internal Error</h1>
    <hr />
    We are sorry to inform you an internal application error has occured:<br />
    %s
    <p>
    Back to <a href="/index">Homepage</a>
    <hr />
    &copy; 2026 <a href="mailto:krylon@gmx.net">Benjamin Walkenhorst</a>
  </body>
</html>
`

	w.Header().Set("Cache-Control", noCache)
	srv.log.Printf("[ERROR] %s\n", msg)

	output := fmt.Sprintf(html, msg)
	w.WriteHeader(500)
	_, _ = w.Write([]byte(output)) // nolint: gosec
} // func (srv *Server) sendErrorMessage(w http.ResponseWriter, msg string)
