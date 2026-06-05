// /home/krylon/go/src/github.com/blicero/hertz/database/database.go
// -*- mode: go; coding: utf-8; -*-
// Created on 01. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-05 21:12:38 krylon>

// Package database implements data persistence.
package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/blicero/hertz/common"
	"github.com/blicero/hertz/logdomain"
	"github.com/blicero/hertz/model"
	"github.com/blicero/krylib"
	"github.com/tidwall/buntdb"
)

var openLock sync.Mutex

// Database wraps the database
type Database struct {
	log  *log.Logger
	path string
	db   *buntdb.DB
}

// Open returns a fresh database instance.
func Open(path string) (*Database, error) {
	var (
		err   error
		exist bool
		db    = &Database{path: path}
	)

	openLock.Lock()
	defer openLock.Unlock()

	if exist, err = krylib.Fexists(path); err != nil {
		return nil, err
	}

	if db.log, err = common.GetLogger(logdomain.Database); err != nil {
		return nil, err
	} else if db.db, err = buntdb.Open(path); err != nil {
		db.log.Printf("[CRITICAL] Failed to open database at %s: %s\n",
			path,
			err.Error())
	}

	if !exist {
		if err = db.initialize(); err != nil {
			return nil, err
		}
	}

	return db, nil
} // func Open(path string) (*Database, error)

func (db *Database) initialize() error {
	db.db.CreateIndex("time_idx", "rec:*", buntdb.IndexJSON("timestamp"))
	db.db.CreateIndex("remote_idx", "remote:*", buntdb.IndexJSON("timestamp"))
	db.db.CreateIndex("client_idx", "client:*", buntdb.IndexInt)
	db.db.Update(func(tx *buntdb.Tx) error {
		tx.Set("id", "0", nil)
		return nil
	})
	return nil
} // func (db *Database) initialize() error

func (db *Database) getID(tx *buntdb.Tx) (int64, error) {
	var (
		err   error
		idStr string
		id    int64
	)

	if idStr, err = tx.Get("id", true); err != nil {
		db.log.Printf("[CRITICAL] Failed to get ID counter: %s\n",
			err.Error())
		return 0, err
	} else if id, err = strconv.ParseInt(idStr, 10, 64); err != nil {
		db.log.Printf("[CRITICAL] Failed to parse ID counter %q: %s\n",
			idStr,
			err.Error())
		return 0, err
	}

	id++
	idStr = strconv.FormatInt(id, 10)

	if _, _, err = tx.Set("id", idStr, nil); err != nil {
		db.log.Printf("[CRITICAL] Failed to update ID counter: %s\n",
			err.Error())
		return 0, err
	}

	return id, nil
} // func (db *Database) getID() (int64, error)

// Close closes the Database.
func (db *Database) Close() {
	db.db.Close()
	db.db = nil
} // func (db *Database) Close()

// RecordAdd adds a Record to the database.
func (db *Database) RecordAdd(rec *model.Record) error {
	var (
		err error
	)

	if err = db.db.Update(func(tx *buntdb.Tx) error {
		var (
			ex               error
			id               int64
			buf              []byte
			jstr, kstr, prev string
			replace          bool
		)

		if id, ex = db.getID(tx); ex != nil {
			return ex
		} else if id == 0 {
			ex = errors.New("getID() returned 0")
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return ex
		}

		rec.ID = id
		kstr = fmt.Sprintf("rec:%d", rec.Timestamp.Unix())

		if buf, ex = json.Marshal(rec); ex != nil {
			db.log.Printf("[ERROR] Cannot serialize Record: %s\n",
				ex.Error())
			return ex
		}

		jstr = string(buf)

		if prev, replace, ex = tx.Set(kstr, jstr, nil); ex != nil {
			db.log.Printf("[ERROR] Failed to save Record: %s\n",
				ex.Error())
			return ex
		} else if replace {
			ex = fmt.Errorf("Key %s already exists, value is %q",
				kstr,
				prev)
			db.log.Printf("[CRITICAL] %s\n", ex.Error())
			return ex
		}

		return nil
	}); err != nil {
		db.log.Printf("[ERROR] Failed to add Record: %s\n",
			err.Error())
		return err
	} else if common.Debug {
		db.log.Printf("[DEBUG] Saved Record %d (%s): %#v\n",
			rec.ID,
			rec.Timestamp.Format(common.TimestampFormat),
			rec.Freq)
	}

	return nil
} // func (db *Database) RecordAdd(rec *model.FreqRecord) error

// RecordAddRemote adds a Record from a remote data collector
func (db *Database) RecordAddRemote(host string, rec *model.Record) error {
	var (
		err error
	)

	if err = db.db.Update(func(tx *buntdb.Tx) error {
		var (
			ex               error
			id               int64
			buf              []byte
			jstr, kstr, prev string
			replace          bool
		)

		if id, ex = db.getID(tx); ex != nil {
			return ex
		} else if id == 0 {
			ex = errors.New("getID() returned 0")
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return ex
		}

		rec.ID = id
		kstr = fmt.Sprintf("remote:%s:%d",
			host,
			rec.Timestamp.Unix())

		if buf, ex = json.Marshal(rec); ex != nil {
			db.log.Printf("[ERROR] Cannot serialize Record: %s\n",
				ex.Error())
			return ex
		}

		jstr = string(buf)

		if prev, replace, ex = tx.Set(kstr, jstr, nil); ex != nil {
			db.log.Printf("[ERROR] Failed to save Record: %s\n",
				ex.Error())
			return ex
		} else if replace {
			ex = fmt.Errorf("Key %s already exists, value is %q",
				kstr,
				prev)
			db.log.Printf("[CRITICAL] %s\n", ex.Error())
			return ex
		}

		var ckey = fmt.Sprintf("client:%s", host)
		var cval = strconv.FormatInt(rec.Timestamp.Unix(), 10)

		if _, _, ex = tx.Set(ckey, cval, nil); ex != nil {
			db.log.Printf("[ERROR] Failed to update timestamp for client %s: %s\n",
				host,
				err.Error())
			return ex
		}

		return nil
	}); err != nil {
		db.log.Printf("[ERROR] Failed to add Record: %s\n",
			err.Error())
		return err
	} else if common.Debug {
		db.log.Printf("[DEBUG] Saved Record %d from %s (%s): %#v\n",
			rec.ID,
			host,
			rec.Timestamp.Format(common.TimestampFormat),
			rec.Freq)
	}

	return nil
} // func (db *Database) RecordAddRemote(host string, rec *model.FreqRecord) error

// RecordGet loads all Records whose timestamp is greater or equal to the given
// time value.
func (db *Database) RecordGet(begin time.Time) ([]*model.Record, error) {
	var (
		err     error
		records []*model.Record
		minKey  string
	)

	records = make([]*model.Record, 0, 32)
	minKey = fmt.Sprintf("rec:%d", begin.Unix())

	if err = db.db.View(func(tx *buntdb.Tx) error {
		var (
			ex error
		)

		if ex = tx.AscendGreaterOrEqual("", minKey, func(key, val string) bool {
			var (
				ex1 error
				buf []byte
				rec = new(model.Record)
			)

			buf = []byte(val)

			if ex1 = json.Unmarshal(buf, rec); ex1 != nil {
				db.log.Printf("[ERROR] Failed to unmarshal record %s: %s\n",
					key,
					ex1.Error())
				return false
			}

			records = append(records, rec)
			return true
		}); ex != nil {
			return ex
		}

		return nil
	}); err != nil {
		db.log.Printf("[ERROR] Failed to fetch records after %s: %s\n",
			begin.Format(common.TimestampFormat),
			err.Error())
		return nil, err
	}

	return records, nil
} // func (db *Database) RecordGet(begin time.Time) ([]*model.FreqRecord, error)

// ClientRegister adds a Client to the Database.
func (db *Database) ClientRegister(name string) (time.Time, error) {
	var (
		err    error
		tstamp = time.Now()
		ckey   = fmt.Sprintf("client:%s", name)
		cval   = strconv.FormatInt(tstamp.Unix(), 10)
	)

	if err = db.db.Update(func(tx *buntdb.Tx) error {
		var (
			ex error
		)

		if _, _, ex = tx.Set(ckey, cval, nil); ex != nil {
			return ex
		}

		return nil
	}); err != nil {
		db.log.Printf("[ERROR] Failed to register Client %s: %s\n",
			name,
			err.Error())
	}

	return tstamp.Truncate(time.Second), err
} // func (db *Database) ClientRegister(name string) (time.Time, error)

// ClientGet looks for a Client's timestamp in the Database.
func (db *Database) ClientGet(name string) (time.Time, error) {
	var (
		err    error
		tstamp time.Time
		key    = fmt.Sprintf("client:%s", name)
	)

	if err = db.db.View(func(tx *buntdb.Tx) error {
		var (
			ex  error
			val string
			n   int64
		)

		if val, ex = tx.Get(key); ex != nil {
			return ex
		} else if n, ex = strconv.ParseInt(val, 10, 64); ex != nil {
			return ex
		}

		tstamp = time.Unix(n, 0)
		return nil
	}); err != nil {
		db.log.Printf("[ERROR] Failed to lookup client %s: %s\n",
			name,
			err.Error())
		return tstamp, err
	}

	return tstamp, nil
} // func (db *Database) ClientGet(name string) (time.Time, error)
