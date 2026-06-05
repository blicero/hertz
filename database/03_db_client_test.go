// /home/krylon/go/src/github.com/blicero/hertz/database/03_db_client_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 05. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-05 20:59:09 krylon>

package database

import (
	"fmt"
	"testing"
	"time"

	"github.com/blicero/hertz/common"
)

const clientCnt = 10

var clients map[string]time.Time

func TestClientAdd(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	clients = make(map[string]time.Time, clientCnt)

	for i := range clientCnt {
		var (
			err   error
			stamp time.Time
			name  string
		)

		name = fmt.Sprintf("client%03d", i)

		if stamp, err = tdb.ClientRegister(name); err != nil {
			t.Fatalf("Registering client %s failed: %s",
				name,
				err.Error())
		}

		clients[name] = stamp
	}
} // func TestClientAdd(t *testing.T)

func TestClientGet(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	for name, s1 := range clients {
		var (
			err error
			s2  time.Time
		)

		if s2, err = tdb.ClientGet(name); err != nil {
			t.Fatalf("Failed to lookup Client %s: %s",
				name,
				err.Error())
		} else if !s1.Equal(s2) {
			t.Fatalf("Unexpected timestamp for Client %s: %s (expected %s)",
				name,
				s2.Format(common.TimestampFormat),
				s1.Format(common.TimestampFormat))
		}
	}
} // func TestClientGet(t *testing.T)
