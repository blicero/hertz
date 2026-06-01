// /home/krylon/go/src/github.com/blicero/hertz/collect/collect_obsd.go
// -*- mode: go; coding: utf-8; -*-
// Created on 01. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-01 13:34:10 krylon>

//go:build openbsd

package collect

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

const (
	ctlPath = "/sbin/sysctl"
	ctlFreq = "hw.cpuspeed"
)

func (p *Probe) getFreq() ([]int64, error) {
	var (
		err    error
		sfreq  string
		freq   int64
		result []int64
	)

	if sfreq, err = p.runSysCtl(ctlFreq); err != nil {
		p.log.Printf("[ERROR] Failed to run sysctl %s: %s\n",
			ctlFreq,
			err.Error())
		return nil, err
	} else if freq, err = strconv.ParseInt(sfreq, 10, 64); err != nil {
		p.log.Printf("[ERROR] Cannot parse CPU frequency %q: %s\n",
			sfreq,
			err.Error())
		return nil, err
	}

	result = make([]int64, runtime.NumCPU())

	for i := range len(result) {
		result[i] = freq
	}

	return result, nil
} // func (p *Probe) getFreq() ([]int64, error)

func (p *Probe) runSysCtl(ctl string) (string, error) {
	var (
		err    error
		cmd    *exec.Cmd
		idx    int
		output []byte
		str    string
	)

	cmd = exec.Command(ctlPath, ctl)

	if output, err = cmd.Output(); err != nil {
		p.log.Printf("[ERROR] Cannot run sysctl(8) - %s\n",
			err.Error())
		return "", err
	}

	str = string(output)

	if idx = strings.Index(str, "="); idx == -1 {
		err = fmt.Errorf("cannot process output of sysctl(8): %s\n",
			output)
		p.log.Printf("[ERROR] %s\n", err.Error())
		return "", err

	}

	return string(output[idx+1 : len(output)-1]), err
} // func (p *Probe) runSysCtl(ctl string) (string, error)
