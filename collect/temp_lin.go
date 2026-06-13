// /home/krylon/go/src/github.com/blicero/hertz/collect/temp_lin.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-13 13:13:20 krylon>

//go:build linux

package collect

import (
	"github.com/blicero/hertz/common"
	"github.com/ssimunic/gosensors"
)

func (p *Probe) getTemp() int64 {
	var (
		err  error
		sens *gosensors.Sensors
		temp int64 = invalidTemp
	)

	if sens, err = gosensors.NewFromSystem(); err != nil {
		p.log.Printf("[ERROR] Failed to collect sensor data: %s\n",
			err.Error())
		return invalidTemp
	}

	if common.Debug {
		p.log.Printf("[DEBUG] Got sensor data: %s\n",
			sens)
	}

	return temp
} // func (p *Probe) getTemp() int64
