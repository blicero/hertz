// /home/krylon/go/src/github.com/blicero/hertz/config/config.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-15 12:47:12 krylon>

// Package config handles reading the configuration file.
package config

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/blicero/krylib"
)

const defaultConf = `
# Time-stamp: <>

[Global]
Debug = true

[Web]
Address = "[::]:7666"

[Collect]
Temperature = ""

[Collect.Interval]
Collect = 15
Transmit = 60
`

// Global is the global section of the config file.
type Global struct {
	Debug bool
}

// Web configures the Web server.
type Web struct {
	Address string
}

// Interval defines the intervals for data collection and transmission.
type Interval struct {
	Collect  int64
	Transmit int64
}

// Collect configures data collection.
type Collect struct {
	Temperature string
	Interval    Interval
}

// Config represents the configuration of the app.
type Config struct {
	Global  Global
	Web     Web
	Collect Collect
}

// Equal returns true if other is a Config value with equal values.
func (cfg *Config) Equal(other any) bool {
	var (
		c2 *Config
		ok bool
	)

	if c2, ok = other.(*Config); !ok {
		return false
	}

	return cfg.Global.Debug == c2.Global.Debug &&
		cfg.Web.Address == c2.Web.Address &&
		cfg.Collect.Temperature == c2.Collect.Temperature &&
		cfg.Collect.Interval.Collect == c2.Collect.Interval.Collect &&
		cfg.Collect.Interval.Transmit == c2.Collect.Interval.Transmit
} // func (cfg *Config) Equal(other any) bool

// Read attempts to read the configuration from the given file.
func Read(path string) (*Config, error) {
	var (
		err    error
		exists bool
		fh     *os.File
		buf    []byte
		cfg    = new(Config)
	)

	if exists, err = krylib.Fexists(path); err != nil {
		return nil, err
	} else if !exists {
		if err = writeDefaultCfg(path); err != nil {
			return nil, err
		}
	}

	if fh, err = os.Open(path); err != nil {
		return nil, err
	}

	defer fh.Close()

	if buf, err = io.ReadAll(fh); err != nil {
		return nil, err
	} else if err = toml.Unmarshal(buf, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
} // func Read(path string) (*Config, error)

func writeDefaultCfg(path string) error {
	var (
		err error
		fh  *os.File
	)

	if fh, err = os.Create(path); err != nil {
		return err
	}

	defer fh.Close()

	if _, err = fh.Write([]byte(defaultConf)); err != nil {
		return err
	}

	return nil
} // func writeDefaultCfg(path string) error
