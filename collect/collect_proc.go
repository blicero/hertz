// /home/krylon/go/src/github.com/blicero/hertz/collect/collect_proc.go
// -*- mode: go; coding: utf-8; -*-
// Created on 30. 05. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-01 14:37:23 krylon>

//go:build linux || freebsd

package collect

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
)

// Sample line:
// cpu MHz		: 1999.970

var freqPat = regexp.MustCompile(`(?ims)^cpu mhz\s+:\s+([0-9.]+)\s*$`)

func (p *Probe) getFreq() ([]int64, error) {
	var (
		err     error
		path    string
		content []byte
	)

	switch runtime.GOOS {
	case "linux":
		path = linuxCPUInfo
	case "freebsd":
		path = freeBSDCpuInfo
	default:
		err = fmt.Errorf("reading /proc/cpuinfo on %s is not supported",
			runtime.GOOS)
		p.log.Printf("[CANTHAPPEN] %s\n",
			err.Error())
		return nil, err
	}

	if content, err = os.ReadFile(path); err != nil {
		p.log.Printf("[ERROR] Cannot read %s: %s\n",
			path,
			err.Error())
		return nil, err
	}

	var info = freqPat.FindAllSubmatch(content, -1)

	// if info == nil {
	// 	return nil, nil
	// }
	if len(info) == 0 {
		return p.getFreqAlt()
	}

	var rec = make([]int64, 0, len(info))

	for _, i := range info {
		var (
			f   float64
			mhz = i[1]
		)

		if f, err = strconv.ParseFloat(string(mhz), 64); err != nil {
			p.log.Printf("[CANTHAPPEN] Cannot parse frequency %q: %s\n",
				mhz,
				err.Error())
			continue
		}

		rec = append(rec, round(f))
	}

	return rec, nil
} // func (p *Probe) getFreq() ([]int64, error)

func (p *Probe) getFreqAlt() ([]int64, error) {
	var (
		err  error
		raw  []byte
		freq int64
		n    = runtime.NumCPU()
	)

	// 1200000 !KHz

	if raw, err = os.ReadFile(linuxSysfsFreq); err != nil {
		p.log.Printf("[ERROR] Cannot open %s: %s\n",
			linuxSysfsFreq,
			err.Error())
		return nil, err
	} else if freq, err = strconv.ParseInt(string(raw[:len(raw)-1]), 10, 64); err != nil {
		p.log.Printf("[ERROR] Cannot parse %s (%q): %s\n",
			linuxSysfsFreq,
			raw,
			err.Error())
		return nil, err
	}

	freq /= 1000
	var data = make([]int64, n)

	for i := range n {
		data[i] = freq
	}

	return data, nil
} // func (p *Probe) getFreqAlt() ([]int64, error)
