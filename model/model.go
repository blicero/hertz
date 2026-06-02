// /home/krylon/go/src/github.com/blicero/hertz/model/model.go
// -*- mode: go; coding: utf-8; -*-
// Created on 30. 05. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-02 10:55:22 krylon>

// Package model provides data types used throughout the application.
package model

import "time"

// FreqRecord contains the CPU frequency at a given point in time.
type FreqRecord struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Freq      []int64   `json:"freq"`
}
