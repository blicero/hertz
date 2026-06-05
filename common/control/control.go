// /home/krylon/go/src/github.com/blicero/hertz/common/control/control.go
// -*- mode: go; coding: utf-8; -*-
// Created on 05. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-05 15:28:17 krylon>

//go:generate stringer -type=Command

package control

// Command identifies the type of Message sent.
type Command uint8

const (
	Start Command = iota
	Stop
	Refresh
)

// Message is a control message.
type Message struct {
	cmd     Command
	payload string
}
