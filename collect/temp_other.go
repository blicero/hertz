// /home/krylon/go/src/github.com/blicero/hertz/collect/temp_other.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-15 13:22:29 krylon>

//go:build !linux

package collect

func (p *Probe) getTemp() (int64, error) {
	p.log.Println("[INFO] getTemp is not implemented on this platform")
	return invalidTemp, nil
} // func (p *Probe) getTemp() (int64, error)
