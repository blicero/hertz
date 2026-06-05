// /home/krylon/go/src/github.com/blicero/hertz/database/01_db_create_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 02. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-05 20:00:43 krylon>

package database

import (
	"testing"

	"github.com/blicero/hertz/common"
)

func TestDBCreate(t *testing.T) {
	var err error

	if tdb, err = Open(common.DbPath); err != nil {
		tdb = nil
		t.Fatalf("Failed to open Database %s: %s\n",
			common.DbPath,
			err.Error())
	}
} // func TestDBCreate(t *testing.T)
