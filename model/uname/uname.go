// /home/krylon/go/src/github.com/blicero/hertz/model/uname/uname.go
// -*- mode: go; coding: utf-8; -*-
// Created on 30. 05. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-05-30 11:53:49 krylon>

// Package uname provides constants to name operating systems.
package uname

//go:generate stringer -type=ID

// ID identifies an operating system.
type ID uint8

const (
	Linux ID = iota
	FreeBSD
	OpenBSD
	NetBSD
)
