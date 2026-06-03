// /home/krylon/go/src/github.com/blicero/hertz/database/database.go
// -*- mode: go; coding: utf-8; -*-
// Created on 01. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-03 12:12:40 krylon>

// Package database implements data persistence.
package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"

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

// RecordAdd adds a Record to the database.
func (db *Database) RecordAdd(rec *model.FreqRecord) error {
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
		kstr = fmt.Sprintf("rec:%d", id)

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
func (db *Database) RecordAddRemote(host string, rec *model.FreqRecord) error {
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
		kstr = fmt.Sprintf("remote:%s:%d", host, id)

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
		db.log.Printf("[DEBUG] Saved Record %d from %s (%s): %#v\n",
			rec.ID,
			host,
			rec.Timestamp.Format(common.TimestampFormat),
			rec.Freq)
	}

	return nil
} // func (db *Database) RecordAddRemote(host string, rec *model.FreqRecord) error
