// /home/krylon/go/src/github.com/blicero/hertz/collect/temp_obsd.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-15 15:27:23 krylon>

//go:build openbsd

package collect

import (
	"fmt"
	"regexp"
	"strconv"
)

const ctlTemp = "hw.sensors.cpu0.temp0"

// Sample: hw.sensors.cpu0.temp0=78.00 degC
var tempPat = regexp.MustCompile(`(\d+[.]\d+) degC`)

func (p *Probe) getTemp() (int64, error) {
	var (
		err     error
		output  string
		match   [][]string
		rawTemp float64
		temp    int64
	)

	if output, err = p.runSysCtl(ctlTemp); err != nil {
		p.log.Printf("[ERROR] Failed to run sysctl %s: %s\n",
			ctlFreq,
			err.Error())
		return invalidTemp, err
	} else if match = tempPat.FindAllStringSubmatch(output, -1); match == nil {
		err = fmt.Errorf("Cannot parse output of sysctl %s: %s",
			ctlTemp,
			output)
		p.log.Printf("[ERROR] %s\n",
			err.Error())
		return invalidTemp, err
	} else if len(match) < 1 {
		err = fmt.Errorf("Cannot parse output of sysctl %s: %s",
			ctlTemp,
			output)
		p.log.Printf("[ERROR] %s\n",
			err.Error())
		return invalidTemp, err
	} else if rawTemp, err = strconv.ParseFloat(match[0][1], 64); err != nil {
		p.log.Printf("[ERROR] Cannot parse temperature %q: %s\n",
			match[0][1],
			err.Error())
		return invalidTemp, err
	}

	temp = round(rawTemp)

	return temp, nil
} // func (p *Probe) getTemp() (int64, error)
