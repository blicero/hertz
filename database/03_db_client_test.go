// /home/krylon/go/src/github.com/blicero/hertz/database/03_db_client_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 05. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-05 21:50:10 krylon>

package database

import (
	"fmt"
	"testing"
	"time"

	"github.com/blicero/hertz/common"
)

const clientCnt = 10

var clients map[string]bool

func TestClientAdd(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	clients = make(map[string]bool, clientCnt)

	for i := range clientCnt {
		var (
			err  error
			name string
		)

		name = fmt.Sprintf("client%03d", i)

		if err = tdb.ClientRegister(name); err != nil {
			t.Fatalf("Registering client %s failed: %s",
				name,
				err.Error())
		}

		clients[name] = true
	}
} // func TestClientAdd(t *testing.T)

func TestClientGet(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	var tzero = time.Unix(0, 0)

	for name := range clients {
		var (
			err   error
			stamp time.Time
		)

		if stamp, err = tdb.ClientGet(name); err != nil {
			t.Fatalf("Failed to lookup Client %s: %s",
				name,
				err.Error())
		} else if !stamp.Equal(tzero) {
			t.Fatalf("Unexpected timestamp for Client %s: %s (expected %s)",
				name,
				stamp.Format(common.TimestampFormat),
				tzero.Format(common.TimestampFormat))
		}
	}
} // func TestClientGet(t *testing.T)
