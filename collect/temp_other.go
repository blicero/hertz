// /home/krylon/go/src/github.com/blicero/hertz/collect/temp_other.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-13 13:09:06 krylon>

//go:build !linux

package collect

func (p *Probe) getTemp() int64 {
	return invalidTemp
} // func (p *Probe) getTemp() int64
