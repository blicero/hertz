// /home/krylon/go/src/github.com/blicero/hertz/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 30. 05. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-09 10:44:23 krylon>

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/blicero/hertz/common"
	"github.com/blicero/hertz/discover"
	"github.com/blicero/hertz/monitor"
	"github.com/blicero/hertz/web"
)

// XXX Set this to a more reasonable value after debugging is done!
const defaultInterval = 30

func main() {
	fmt.Printf("%s %s, built on %s\n",
		common.AppName,
		common.Version,
		common.BuildStamp.Format(common.TimestampFormat))

	var (
		err             error
		collectInterval int64
		xmitInterval    int64
		webAddr         = fmt.Sprintf("[::]:%d", common.WebPort)
		runWeb, runMon  bool
		mode            string
		xp              *discover.Explorer
		ticker          *time.Ticker
		sigQ            chan os.Signal
		mon             *monitor.Monitor
		srv             *web.Server
	)

	flag.Int64Var(
		&collectInterval,
		"cinterval",
		defaultInterval,
		"Interval (in seconds) between data collections")

	flag.Int64Var(
		&xmitInterval,
		"xinterval",
		int64(common.LiveTimeout.Seconds()),
		"Interval (in seconds) between data transmissions",
	)

	flag.StringVar(
		&webAddr,
		"addr",
		webAddr,
		"IP Adress for the web server to listen on")

	flag.BoolVar(
		&runWeb,
		"web",
		false,
		"Open the web interface?",
	)

	flag.BoolVar(
		&runMon,
		"mon",
		false,
		"Run the Monitor",
	)

	flag.Parse()

	if !(runMon || runWeb) {
		fmt.Fprint(
			os.Stderr,
			"At least one flag of -mon or -web must be given!",
		)
		os.Exit(1)
	}

	if runMon {
		mode = "agent"
		if mon, err = monitor.Create(collectInterval); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Failed to create Monitor: %s\n",
				err.Error())
			os.Exit(1)
		}

		mon.Start()
	}

	if runWeb {
		mode = "server"
		if srv, err = web.Create(webAddr); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Creating web server failed: %s\n",
				err.Error())
			os.Exit(2)
		}

		go srv.Run()
	}

	if xp, err = discover.Create(mode, time.Duration(xmitInterval)*time.Second); err != nil {
		fmt.Fprint(
			os.Stderr,
			"Failed to initialize peer discovery in %s mode: %s\n",
			mode,
			err.Error(),
		)
	}

	ticker = time.NewTicker(common.ActiveTimeout)
	defer ticker.Stop()

	sigQ = make(chan os.Signal, 1)
	signal.Notify(sigQ, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			continue
		case s := <-sigQ:
			fmt.Fprintf(
				os.Stderr,
				"Caught signal: %s\n",
				s)

			if mon != nil {
				mon.Stop()
			}

			if srv != nil {
				srv.Stop()
			}

			if xp != nil {
				xp.Shutdown()
			}

			return
		}
	}
}
