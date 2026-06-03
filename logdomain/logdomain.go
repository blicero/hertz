// /home/krylon/go/src/github.com/blicero/hertz/logdomain/logdomain.go
// -*- mode: go; coding: utf-8; -*-
// Created on 30. 05. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-03 11:14:07 krylon>

// Package logdomain provides constants to identify different parts of
// the application.
package logdomain

//go:generate stringer -type=ID

// ID identifies a part of the application that wants to write messages to the log.
type ID uint8

const (
	Collect ID = iota
	Database
	Monitor
	Web
)

// All returns a slice of all valid IDs.
func All() []ID {
	return []ID{
		Collect,
		Database,
		Monitor,
		Web,
	}
}
