// /home/krylon/go/src/github.com/blicero/hertz/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 30. 05. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-16 14:14:07 krylon>

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/blicero/hertz/collect"
	"github.com/blicero/hertz/common"
	"github.com/blicero/hertz/config"
	"github.com/blicero/hertz/discover"
	"github.com/blicero/hertz/logdomain"
	"github.com/blicero/hertz/monitor"
	"github.com/blicero/hertz/web"
	"github.com/hashicorp/logutils"
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
		cfg             *config.Config
	)

	if err = common.InitApp(); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Failed to initialized application environment: %s\n",
			err.Error(),
		)
		os.Exit(1)
	} else if cfg, err = config.Read(common.CfgPath); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Failed to read config from %s: %s\n",
			common.CfgPath,
			err.Error())
		os.Exit(1)
	}

	common.PackageLevels[logdomain.Collect] = logutils.LogLevel(cfg.Loglevel.Collect)
	common.PackageLevels[logdomain.Database] = logutils.LogLevel(cfg.Loglevel.Database)
	common.PackageLevels[logdomain.DBPool] = logutils.LogLevel(cfg.Loglevel.DBPool)
	common.PackageLevels[logdomain.Monitor] = logutils.LogLevel(cfg.Loglevel.Monitor)
	common.PackageLevels[logdomain.Web] = logutils.LogLevel(cfg.Loglevel.Web)
	common.PackageLevels[logdomain.Discover] = logutils.LogLevel(cfg.Loglevel.Discover)
	common.PackageLevels[logdomain.Client] = logutils.LogLevel(cfg.Loglevel.Client)

	webAddr = cfg.Web.Address
	common.Debug = cfg.Global.Debug
	collect.TemperaturePath = cfg.Collect.Temperature

	flag.Int64Var(
		&collectInterval,
		"cinterval",
		cfg.Collect.Interval.Collect,
		"Interval (in seconds) between data collections")

	flag.Int64Var(
		&xmitInterval,
		"xinterval",
		cfg.Collect.Interval.Transmit,
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
		fmt.Fprintf(
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
