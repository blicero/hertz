// /home/krylon/go/src/github.com/blicero/hertz/database/database.go
// -*- mode: go; coding: utf-8; -*-
// Created on 01. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-01 16:13:42 krylon>

// Package database implements data persistence.
package database

import (
	"log"

	"github.com/blicero/hertz/common"
	"github.com/blicero/hertz/logdomain"
	"github.com/tidwall/buntdb"
)

// Database wraps the database
type Database struct {
	log  *log.Logger
	path string
	db   *buntdb.DB
}

// Open returns a fresh database instance.
func Open(path string) (*Database, error) {
	var (
		err error
		db  = &Database{path: path}
	)

	if db.log, err = common.GetLogger(logdomain.Database); err != nil {
		return nil, err
	} else if db.db, err = buntdb.Open(path); err != nil {
		db.log.Printf("[CRITICAL] Failed to open database at %s: %s\n",
			path,
			err.Error())
	}

	return db, nil
} // func Open(path string) (*Database, error)
