// /home/krylon/go/src/github.com/blicero/hertz/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 30. 05. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-02 12:47:40 krylon>

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/blicero/hertz/common"
	"github.com/blicero/hertz/monitor"
)

// XXX Set this to a more reasonable value after debugging is done!
const defaultInterval = 30

func main() {
	fmt.Printf("%s %s, built on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	var (
		err      error
		interval int64
		ticker   *time.Ticker
		sigQ     chan os.Signal
		mon      *monitor.Monitor
	)

	flag.Int64Var(
		&interval,
		"interval",
		defaultInterval,
		"Interval (in seconds) between data collections")

	flag.Parse()

	if mon, err = monitor.Create(interval); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Failed to create Monitor: %s\n",
			err.Error())
		os.Exit(1)
	}

	mon.Start()

	ticker = time.NewTicker(common.ActiveTimeout)
	defer ticker.Stop()

	sigQ = make(chan os.Signal, 1)
	signal.Notify(sigQ, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			if !mon.IsActive() {
				return
			}
		case s := <-sigQ:
			fmt.Fprintf(
				os.Stderr,
				"Caught signal: %s\n",
				s)
			mon.Stop()
			return
		}
	}
}
