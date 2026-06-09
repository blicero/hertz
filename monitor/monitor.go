// /home/krylon/go/src/github.com/blicero/hertz/monitor/monitor.go
// -*- mode: go; coding: utf-8; -*-
// Created on 02. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-09 11:29:57 krylon>

// Package monitor implements the process of collecting and storing data
// in a regular manner.
package monitor

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blicero/hertz/collect"
	"github.com/blicero/hertz/common"
	"github.com/blicero/hertz/logdomain"
	"github.com/blicero/hertz/model"
)

// Monitor collects data and stores it to the database.
type Monitor struct {
	log       *log.Logger
	probe     *collect.Probe
	active    atomic.Bool
	wg        sync.WaitGroup
	cinterval time.Duration
}

// Create creates a new Monitor that gathers new data every <tickSeconds> seconds.
func Create(tickSeconds int64) (*Monitor, error) {
	var (
		err error
		mon = &Monitor{
			cinterval: time.Second * time.Duration(tickSeconds),
		}
	)

	if mon.log, err = common.GetLogger(logdomain.Monitor); err != nil {
		return nil, err
	} else if mon.probe, err = collect.New(); err != nil {
		mon.log.Printf("[CRITICAL] Cannot open Probe: %s\n",
			err.Error())
		return nil, err
	}

	return mon, nil
} // func Create() (*Monitor, error)

// Start sets the Monitor to run.
func (mon *Monitor) Start() {
	mon.active.Store(true)
	mon.wg.Go(mon.process)
} // func (mon *Monitor) Start()

// Stop tells the Monitor to stop collecting data by clearing its active flag.
func (mon *Monitor) Stop() {
	mon.active.Store(false)
	mon.wg.Wait()
} // func (mon *Monitor) Stop()

// IsActive returns the Monitor's active flag.
func (mon *Monitor) IsActive() bool {
	return mon.active.Load()
} // func (*mon Monitor) IsActive() bool

func (mon *Monitor) process() {
	var (
		err    error
		rec    *model.Record
		ticker *time.Ticker
	)

	mon.log.Printf("[TRACE] Starting to collect data every %s\n",
		mon.cinterval)
	defer mon.log.Println("[TRACE] Terminating Monitor process")

	ticker = time.NewTicker(mon.cinterval)
	defer ticker.Stop()

	for mon.active.Load() {
		if rec, err = mon.probe.Collect(); err != nil {
			mon.log.Printf("[ERROR] Cannot gather data: %s\n",
				err.Error())
		} else if err = mon.recordStore(rec); err != nil {
			mon.log.Printf("[ERROR] Failed to save record to DB: %s\n",
				err.Error())
		}

		<-ticker.C
	}
} // func (mon *Monitor) process()

func (mon *Monitor) recordStore(rec *model.Record) error {
	var (
		err  error
		path string
		buf  []byte
		fh   *os.File
	)

	if rec == nil {
		err = errors.New("Record is nil")
		mon.log.Printf("[ERROR] %s\n", err.Error())
		return err
	} else {
		mon.log.Printf("[DEBUG] Save one record to file: %#v\n",
			rec)
	}

	path = filepath.Join(
		common.SpoolDir,
		fmt.Sprintf("%016x.json", rec.Timestamp.Unix()),
	)

	if buf, err = json.Marshal(rec); err != nil {
		mon.log.Printf("[ERROR] Cannot serialize Record: %s\n",
			err.Error())
		return err
	} else if fh, err = os.Create(path); err != nil {
		mon.log.Printf("[ERROR] Cannot open %s: %s\n",
			path,
			err.Error())
		return err
	}

	defer fh.Close() // nolint: errcheck

	if _, err = fh.Write(buf); err != nil {
		mon.log.Printf("[ERROR] Failed to write Record to %s: %s\n",
			path,
			err.Error())
		return err
	}

	return nil
} // func (mon *Monitor) recordStore(rec *model.Record) error
