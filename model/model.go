// /home/krylon/go/src/github.com/blicero/hertz/model/model.go
// -*- mode: go; coding: utf-8; -*-
// Created on 30. 05. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-05 21:30:05 krylon>

// Package model provides data types used throughout the application.
package model

import "time"

// Record contains the CPU frequency at a given point in time.
type Record struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Freq      []int64   `json:"freq"`
	UUID      string    `json:"uuid"`
}

// Equal returns true if the Record is equal to other.
func (r *Record) Equal(other *Record) bool {
	if r.ID != other.ID ||
		!r.Timestamp.Equal(other.Timestamp) ||
		r.UUID != other.UUID ||
		len(r.Freq) != len(other.Freq) {
		return false
	}

	for i, f := range r.Freq {
		if f != other.Freq[i] {
			return false
		}
	}

	return true
} // func (r FreqRecord) Equal(other FreqRecord) bool

// SrvResponse is the response sent by the Server to the Client.
type SrvResponse struct {
	Status    bool      `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Payload   string    `json:"payload"`
}
