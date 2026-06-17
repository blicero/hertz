// /home/krylon/go/src/github.com/blicero/hertz/client/client.go
// -*- mode: go; coding: utf-8; -*-
// Created on 05. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-17 11:51:01 krylon>

// Package client handles communication with a Server.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blicero/hertz/common"
	"github.com/blicero/hertz/common/control"
	"github.com/blicero/hertz/database"
	"github.com/blicero/hertz/logdomain"
	"github.com/blicero/hertz/model"
)

// Client wraps the state for communicating with a Server.
type Client struct {
	log      *log.Logger
	srv      string
	hostname string
	interval time.Duration
	active   atomic.Bool
	client   *http.Client
	wg       *sync.WaitGroup
	db       *database.Database
	cmdQ     <-chan control.Message
}

// New creates a new Client that will talk to the Server at the given address.
func New(addr string, interval time.Duration, cmdQ <-chan control.Message) (*Client, error) {
	var (
		err error
		c   = &Client{
			srv:  addr,
			cmdQ: cmdQ,
			wg:   new(sync.WaitGroup),
			client: &http.Client{
				Timeout: time.Millisecond * 2500,
			},
			interval: interval,
		}
	)

	if c.log, err = common.GetLogger(logdomain.Client); err != nil {
		return nil, err
	} else if c.db, err = database.Open(common.DbPath); err != nil {
		c.log.Printf("[CRITICAL] Failed to open database: %s\n",
			err.Error())
		return nil, err
	}

	c.hostname, _ = os.Hostname()

	if idx := strings.Index(c.hostname, "."); idx > 0 {
		c.hostname = c.hostname[:idx]
	}

	return c, nil
} // func New(addr string, cmdQ <-chan control.Message) (*Client, error)

// Start starts the client main loop.
func (c *Client) Start() {
	c.active.Store(true)

	c.wg.Go(c.run)
} // func (c *Client) Start()

// Stop shuts down the client.
func (c *Client) Stop() {
	c.active.Store(false)
}

func (c *Client) run() {
	var (
		err                    error
		timestamp              time.Time
		liveTicker, xmitTicker *time.Ticker
	)

	c.log.Printf("[INFO] Transmitting data to %s every %s\n",
		c.srv,
		c.interval)
	defer c.log.Println("[INFO] Client is quitting")

	if timestamp, err = c.getTimestamp(); err != nil {
		c.log.Printf("[ERROR] Failed to get timestamp from %s: %s\n",
			c.srv,
			err.Error())
		return
	}

	liveTicker = time.NewTicker(common.ActiveTimeout)
	defer liveTicker.Stop()

	xmitTicker = time.NewTicker(c.interval)
	defer xmitTicker.Stop()

	for c.active.Load() {
		var msg control.Message

		select {
		case <-liveTicker.C:
			continue
		case <-xmitTicker.C:
			var t time.Time
			if t, err = c.transmitData(timestamp); err != nil {
				c.log.Printf("[ERROR] Failed to transmit data: %s\n",
					err.Error())
			} else {
				timestamp = t
			}
		case msg = <-c.cmdQ:
			c.log.Printf("[DEBUG] Received control message: %s\n",
				msg.Cmd)
			switch msg.Cmd {
			case control.Start:
			case control.Stop:
				c.Stop()
			case control.Refresh:

			}
		}
	}
} // func (c *Client) run()

func (c *Client) getTimestamp() (time.Time, error) {
	var (
		err       error
		uri       string
		resp      *http.Response
		buf       []byte
		data      model.SrvResponse
		unixStamp int64
		timestamp time.Time
	)

	uri = fmt.Sprintf("%s/ws/get_timestamp/%s",
		c.srv,
		c.hostname)

	if resp, err = c.client.Get(uri); err != nil {
		c.log.Printf("[ERROR] Failed to get timestamp from server: %s\n",
			err.Error())
		return timestamp, err
	} else if resp == nil {
		err = fmt.Errorf("http.Client.Get(%s) return a nil response",
			uri)
		c.log.Printf("[CANTHAPPEN] %s\n",
			err.Error())
		return timestamp, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("unexpected HTTP status code in response: %s",
			resp.Status)
		c.log.Printf("[ERROR] %s\n",
			err.Error())
		return timestamp, err
	} else if buf, err = io.ReadAll(resp.Body); err != nil {
		c.log.Printf("[ERROR] Failed to read response from Server: %s\n",
			err.Error())
		return timestamp, err
	} else if err = json.Unmarshal(buf, &data); err != nil {
		c.log.Printf("[ERROR] Failed to parse response from Server: %s\n%s\n\n",
			err.Error(),
			buf)
		return timestamp, err
	} else if unixStamp, err = strconv.ParseInt(data.Payload, 10, 64); err != nil {
		c.log.Printf("[ERROR] Failed to parse timestamp %q: %s\n",
			data.Payload,
			err.Error())
		return timestamp, err
	}

	timestamp = time.Unix(unixStamp, 0)

	return timestamp, nil
} // func (c *Client) getTimestamp() (time.Time, error)

func (c *Client) transmitData(t time.Time) (time.Time, error) {
	var (
		err      error
		records  []*model.Record
		files    []string
		serial   []byte
		buf      *bytes.Buffer
		res      *http.Response
		reply    model.SrvResponse
		recent   time.Time
		endpoint string
	)

	if records, files, err = c.loadData(t); err != nil {
		c.log.Printf("[ERROR] Failed to get Records after %s: %s\n",
			t.Format(common.TimestampFormat),
			err.Error())
		return t, err
	} else if len(records) == 0 {
		return t, nil
	} else if serial, err = json.Marshal(records); err != nil {
		c.log.Printf("[ERROR] Cannot serialize %d Records: %s\n",
			len(records),
			err.Error())
		return t, err
	}

	c.log.Printf("[TRACE] Transmitting %d Records to %s\n",
		len(records),
		c.srv)

	recent = records[len(records)-1].Timestamp
	buf = bytes.NewBuffer(serial)
	endpoint = fmt.Sprintf("%s/ws/submit_data/%s",
		c.srv,
		c.hostname)

	if res, err = c.client.Post(endpoint, "application/json", buf); err != nil {
		c.log.Printf("[ERROR] Failed to submit data to %s: %s\n",
			c.srv,
			err.Error())
		return t, err
	}

	defer res.Body.Close() // nolint: nilaway

	// nolint: nilaway
	if res.StatusCode != 200 {
		err = fmt.Errorf("HTTP request to submit data to %s failed: %s",
			c.srv,
			res.Status)
		c.log.Printf("[ERROR] %s\n",
			err.Error())
		return t, err
	}

	buf.Reset()
	io.Copy(buf, res.Body) // nolint: errcheck

	var rstamp = strconv.FormatInt(recent.Unix(), 10)

	if err = json.Unmarshal(buf.Bytes(), &reply); err != nil {
		c.log.Printf("[ERROR] Failed to parse response from %s: %s\n%s\n\n",
			c.srv,
			err.Error(),
			buf.Bytes())
		return t, err
	} else if !reply.Status {
		err = fmt.Errorf("server-side error handling request to %s: %s",
			endpoint,
			reply.Message)
		c.log.Printf("[ERROR] %s\n",
			err.Error())
		return t, err
	} else if reply.Payload != rstamp {
		err = fmt.Errorf("Unexpected Payload in response from server: %s (expected %s)",
			reply.Payload,
			rstamp)
		c.log.Printf("[ERROR] %s\n", err.Error())
		return t, err
	} else {
		c.log.Printf("[TRACE] Server acknowledged receipt of data: %s\n",
			reply.Message)
	}

	for _, f := range files {
		var path = filepath.Join(common.SpoolDir, f)

		if err = os.Remove(path); err != nil {
			c.log.Printf("[ERROR] Cannot delete spool file %s: %s\n",
				f,
				err.Error())
		}
	}

	return recent, nil
} // func (c *Client) transmitData(t time.Time) error

func (c *Client) loadData(t time.Time) ([]*model.Record, []string, error) {
	var (
		err   error
		dirh  *os.File
		files []string
		data  []*model.Record
	)

	if dirh, err = os.Open(common.SpoolDir); err != nil {
		c.log.Printf("[CRITICAL] Cannot open spool directory %s: %s\n",
			common.SpoolDir,
			err.Error())
		return nil, nil, err
	}

	defer dirh.Close() // nolint: errcheck

	if files, err = dirh.Readdirnames(-1); err != nil {
		c.log.Printf("[CRITICAL] Cannot read contents of spool directory %s: %s\n",
			common.SpoolDir,
			err.Error())
		return nil, nil, err
	}

	data = make([]*model.Record, 0, len(files))

	for _, f := range files {
		var (
			path, tstr string
			timestamp  int64
			buf        []byte
			rec        = new(model.Record)
		)

		tstr = f[:16]
		if timestamp, err = strconv.ParseInt(tstr, 16, 64); err != nil {
			c.log.Printf("[ERROR] Cannot parse timestamp from filename %q: %s\n",
				path,
				err.Error())
			return nil, nil, err
		} else if time.Unix(timestamp, 0).Before(t) {
			continue
		}

		path = filepath.Join(common.SpoolDir, f)

		if buf, err = os.ReadFile(path); err != nil {
			c.log.Printf("[ERROR] Cannot read %s: %s\n",
				path,
				err.Error())
			return nil, nil, err
		} else if err = json.Unmarshal(buf, rec); err != nil {
			c.log.Printf("[ERROR] Cannot parse Record from %s: %s\n",
				path,
				err.Error())
			return nil, nil, err
		} else if rec == nil {
			c.log.Printf("[ERROR] No error was returned processing %s, but no Record, either\n",
				path)
			continue
		}

		data = append(data, rec)
	}

	return data, files, nil
} // func (c *Client) loadData(t time.Time) ([]*model.Record, error)
