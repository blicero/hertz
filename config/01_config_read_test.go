// /home/krylon/go/src/github.com/blicero/hertz/config/config_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-16 13:54:45 krylon>

package config

import (
	"testing"

	"github.com/blicero/hertz/common"
	"github.com/davecgh/go-spew/spew"
)

func TestReadNoConfig(t *testing.T) {
	var (
		err    error
		cfg    *Config
		expect = Config{
			Global: Global{
				Debug: true,
			},
			Web: Web{
				Address: "[::]:7666",
			},
			Collect: Collect{
				Temperature: "",
				Interval: Interval{
					Collect:  15,
					Transmit: 60,
				},
			},
			Loglevel: Loglevel{
				Collect:  "DEBUG",
				Database: "DEBUG",
				DBPool:   "DEBUG",
				Monitor:  "DEBUG",
				Web:      "DEBUG",
				Discover: "DEBUG",
				Client:   "DEBUG",
			},
		}
	)

	if cfg, err = Read(common.CfgPath); err != nil {
		t.Fatalf("Failed to read %s: %s\n",
			common.CfgPath,
			err.Error())
	} else if cfg == nil {
		t.Fatal("Read() did not return a Config object")
	} else if !cfg.Equal(&expect) {
		t.Fatalf("Read() returned unexpected config:\nExpected: %s\nGot: %s\n",
			spew.Sdump(&expect),
			spew.Sdump(cfg))
	}
} // func TestReadNoConfig(t *testing.T)

func TestReadExampleConfig(t *testing.T) {
	const cfgPath = "testdata/test01.toml"
	var (
		err    error
		cfg    *Config
		expect = Config{
			Global: Global{
				Debug: false,
			},
			Web: Web{
				Address: "[::]:2342",
			},
			Collect: Collect{
				Temperature: "",
				Interval: Interval{
					Collect:  30,
					Transmit: 90,
				},
			},
			Loglevel: Loglevel{
				Collect:  "DEBUG",
				Database: "DEBUG",
				DBPool:   "DEBUG",
				Monitor:  "DEBUG",
				Web:      "DEBUG",
				Discover: "DEBUG",
				Client:   "DEBUG",
			},
		}
	)

	if cfg, err = Read(cfgPath); err != nil {
		t.Fatalf("Failed to read %s: %s\n",
			common.CfgPath,
			err.Error())
	} else if cfg == nil {
		t.Fatal("Read() did not return a Config object")
	} else if !cfg.Equal(&expect) {
		t.Fatalf("Read() returned unexpected config:\nExpected: %s\nGot: %s\n",
			spew.Sdump(&expect),
			spew.Sdump(cfg))
	}
} // func TestReadExampleConfig(t *testing.T)
