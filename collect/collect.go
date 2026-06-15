// /home/krylon/go/src/github.com/blicero/hertz/collect/collect.go
// -*- mode: go; coding: utf-8; -*-
// Created on 30. 05. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-15 13:37:55 krylon>

// Package collect implements collecting CPU frequency data.
package collect

import (
	"log"
	"math"
	"time"

	"github.com/blicero/hertz/common"
	"github.com/blicero/hertz/logdomain"
	"github.com/blicero/hertz/model"
)

// TemperaturePath defines - on Linux systems - what sensor to read the
// temperature from.
var TemperaturePath string

// nolint: unused
const (
	linuxCPUInfo   = "/proc/cpuinfo"
	linuxSysfsFreq = "/sys/devices/system/cpu/cpufreq/policy0/scaling_cur_freq"
	freeBSDCpuInfo = "/linproc/cpuinfo"
	invalidTemp    = -1024
)

// nolint: unused
func round(f float64) int64 {
	return int64(math.Floor(f + 0.5))
}

// Probe gathers CPU frequency data.
type Probe struct {
	log *log.Logger
}

// New creates a fresh Probe
func New() (*Probe, error) {
	var (
		err error
		p   = new(Probe)
	)

	if p.log, err = common.GetLogger(logdomain.Collect); err != nil {
		return nil, err
	}

	return p, nil
} // func New() (*Probe, error)

// Collect attempts to collect data on the current CPU frequency.
func (p *Probe) Collect() (*model.Record, error) {
	var (
		err error
		rec = &model.Record{}
	)

	if rec.Freq, err = p.getFreq(); err != nil {
		p.log.Printf("[ERROR] Failed to collect frequency data: %s\n",
			err.Error())
		return nil, err
	}

	if rec.Temperature, err = p.getTemp(); err != nil {
		p.log.Printf("[ERROR] Failed to read temperature: %s\n",
			err.Error())
		rec.Temperature = invalidTemp
	}
	rec.Timestamp = time.Now()
	return rec, nil
} // func (p *Probe) Collect() (*model.FreqRecord, error)
