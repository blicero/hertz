// /home/krylon/go/src/github.com/blicero/hertz/model/model.go
// -*- mode: go; coding: utf-8; -*-
// Created on 30. 05. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-05-30 11:52:10 krylon>

// Package model provides data types used throughout the application.
package model

import "time"

type FreqRecord struct {
	ID        int64
	Timestamp time.Time
	Freq      []int64
}
