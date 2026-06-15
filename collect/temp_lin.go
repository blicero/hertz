// /home/krylon/go/src/github.com/blicero/hertz/collect/temp_lin.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-15 13:38:17 krylon>

//go:build linux

package collect

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/blicero/hertz/common"
	"github.com/ssimunic/gosensors"
)

// ErrInvalidPath indicates an invalid path for the temperature sensor.
var ErrInvalidPath = errors.New("Path to temperature sensor is invalid")

// sample value: +48.0°C  (high = +100.0°C, crit = +100.0°C)
var tempPat = regexp.MustCompile(`^[-+]?(\d+[.]\d+)°C`)

func (p *Probe) getTemp() (int64, error) {
	var (
		err     error
		sens    *gosensors.Sensors
		temp    int64 = invalidTemp
		path    []string
		value   string
		rawTemp float64
		ok      bool
	)

	if sens, err = gosensors.NewFromSystem(); err != nil {
		p.log.Printf("[ERROR] Failed to collect sensor data: %s\n",
			err.Error())
		return invalidTemp, err
	}

	path = strings.Split(TemperaturePath, "/")

	if len(path) != 2 {
		return invalidTemp, ErrInvalidPath
	}

	for chip := range sens.Chips {
		var match []string

		if chip != path[0] {
			continue
		} else if value, ok = sens.Chips[chip][path[1]]; !ok {
			p.log.Printf("[ERROR] No data was found at %s\n",
				TemperaturePath)
			return invalidTemp, ErrInvalidPath
		} else if match = tempPat.FindAllString(value, -1); match == nil {
			err = fmt.Errorf("cannot parse output from sensor: %s",
				err.Error())
			p.log.Printf("[ERROR] %s\n", err.Error())
			return invalidTemp, err
		} else if rawTemp, err = strconv.ParseFloat(match[1], 64); err != nil {
			p.log.Printf("[ERROR] Cannot parse floating point number %q: %s\n",
				match[1],
				err.Error())
			return invalidTemp, err
		}

		temp = round(rawTemp)
	}

	if common.Debug {
		p.log.Printf("[DEBUG] Got sensor data: %s\n",
			sens)
	}

	return temp, nil
} // func (p *Probe) getTemp() int64
