// /home/krylon/go/src/github.com/blicero/hertz/database/query/query.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-08 12:17:59 krylon>

//go:generate stringer -type=ID

// Package query provides IDs to uniquely reference database queries.
package query

// ID identifies a particular Database query.
type ID uint8

const (
	HostAdd ID = iota
	HostGetByName
	HostGetByID
	HostGetRecent
	HostUpdateLastContact
	HostDelete
	RecordAdd
	RecordGetByHost
	RecordGetByPeriod
	RecordGetByHostPeriod
)
