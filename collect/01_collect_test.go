// /home/krylon/go/src/github.com/blicero/hertz/collect/01_collect_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 01. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-05 20:23:30 krylon>

package collect

import (
	"testing"

	"github.com/blicero/hertz/model"
)

var tp *Probe

func TestProbeCreate(t *testing.T) {
	var err error

	if tp, err = New(); err != nil {
		tp = nil
		t.Fatalf("Failed to create Probe: %s",
			err.Error())
	} else if tp == nil {
		t.Error("New returned nil, but no error")
	}
} // func TestProbeCreate(t *testing.T)

func TestProbeCollect(t *testing.T) {
	if tp == nil {
		t.SkipNow()
	}

	var (
		err error
		rec *model.Record
	)

	if rec, err = tp.Collect(); err != nil {
		t.Fatalf("Failed to collect data: %s",
			err.Error())
	} else if rec == nil {
		t.Error("Collect returned nil, but no error")
	} else if len(rec.Freq) == 0 {
		t.Error("Collect did not return any data")
	}
} // func TestProbeCollect(t *testing.T)
