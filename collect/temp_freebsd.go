// /home/krylon/go/src/github.com/blicero/hertz/collect/temp_freebsd.go
// -*- mode: go; coding: utf-8; -*-
// Created on 16. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-16 12:05:51 krylon>

//go:build freebsd

package collect

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Sample output of ipmitool(1):
// 02-CPU           | 40,000     | degrees C  | ok    | na        | na        | na        | na        | 70,000    | na

const ipmitool = "/usr/local/bin/ipmitool"

var tempPat = regexp.MustCompile(`(?msi:^\d+-CPU\s+\|\s+(\d+[,.]\d+))`)

func (p *Probe) getTemp() (int64, error) {
	var (
		err     error
		output  []byte
		cmd     *exec.Cmd
		match   [][][]byte
		rawTemp float64
		temp    int64
	)

	cmd = exec.Command(ipmitool, "sensor")

	if output, err = cmd.Output(); err != nil {
		p.log.Printf("[ERROR] Failed to call ipmitool(1): %s\n",
			err.Error())
		return invalidTemp, err
	} else if match = tempPat.FindAllSubmatch(output, -1); len(match) == 0 {
		err = fmt.Errorf("Cannot parse output of impitool: %s",
			output)
		p.log.Printf("[ERROR] %s\n",
			err.Error())
		return invalidTemp, err
	}

	var tempStr = string(match[0][1])
	tempStr = strings.Replace(tempStr, ",", ".", 1)

	if rawTemp, err = strconv.ParseFloat(tempStr, 64); err != nil {
		p.log.Printf("[ERROR] Cannot parse temperature %q: %s\n",
			match[0][1],
			err.Error())
		return invalidTemp, err
	}

	temp = round(rawTemp)
	return temp, nil
} // func (p *Probe) getTemp() (int64, error)
