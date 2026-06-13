// /home/krylon/go/src/github.com/blicero/hertz/model/model.go
// -*- mode: go; coding: utf-8; -*-
// Created on 30. 05. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-13 11:07:05 krylon>

// Package model provides data types used throughout the application.
package model

import (
	"time"
)

type Host struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	LastContact time.Time `json:"last_contact"`
}

// IsAlive returns true if the most recent interaction with the Host
// was within twice the usual transmit interval.
func (h *Host) IsAlive() bool {
	return time.Since(h.LastContact) < (time.Minute * 5)
} // func (h *Host) IsAlive() bool

// Record contains the CPU frequency at a given point in time.
type Record struct {
	ID        int64     `json:"id"`
	HostID    int64     `json:"host_id"`
	Timestamp time.Time `json:"timestamp"`
	Freq      []int64   `json:"freq"`
}

// Equal returns true if the Record is equal to other.
func (r *Record) Equal(other *Record) bool {
	if r.ID != other.ID ||
		r.HostID != other.HostID ||
		!r.Timestamp.Equal(other.Timestamp) ||
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
