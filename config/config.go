// /home/krylon/go/src/github.com/blicero/hertz/config/config.go
// -*- mode: go; coding: utf-8; -*-
// Created on 15. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-16 14:21:54 krylon>

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

[Loglevel]
Collect = "DEBUG"
Database = "DEBUG"
DBPool = "DEBUG"
Monitor = "DEBUG"
Web = "DEBUG"
Discover = "DEBUG"
Client = "DEBUG"
`

// Global is the global section of the config file.
type Global struct {
	Debug bool
}

// Equal returns true if other is equal to itself.
func (g *Global) Equal(other any) bool {
	var (
		g2 *Global
		ok bool
	)

	if g2, ok = other.(*Global); !ok {
		return false
	}

	return g.Debug == g2.Debug
} // func (g *Global) Equal(other any) bool

// Web configures the Web server.
type Web struct {
	Address string
}

// Equal returns true if other is equal to itself.
func (w *Web) Equal(other any) bool {
	var (
		x  *Web
		ok bool
	)

	if x, ok = other.(*Web); !ok {
		return false
	}

	return w.Address == x.Address
} // func (w *Web) Equal(other any) bool

// Interval defines the intervals for data collection and transmission.
type Interval struct {
	Collect  int64
	Transmit int64
}

// Equal returns true if other is equal to itself.
func (i *Interval) Equal(other any) bool {
	var (
		i2 *Interval
		ok bool
	)

	if i2, ok = other.(*Interval); !ok {
		return false
	}

	return (i.Collect == i2.Collect) &&
		(i.Transmit == i2.Transmit)
} // func (i *Interval) Equal(other any) bool

// Collect configures data collection.
type Collect struct {
	Temperature string
	Interval    Interval
}

// Equal returns true if other is equal to itself.
func (c *Collect) Equal(other any) bool {
	var (
		c2 *Collect
		ok bool
	)

	if c2, ok = other.(*Collect); !ok {
		return false
	}

	return (c.Temperature == c2.Temperature) &&
		(c.Interval == c2.Interval)
} // func (c *Collect) Equal(other any) bool

// Loglevel configures the minimum log level for the components of
// the application.
type Loglevel struct {
	Collect  string
	Database string
	DBPool   string
	Monitor  string
	Web      string
	Discover string
	Client   string
}

// Equal returns true if other is equal to itself.
func (l *Loglevel) Equal(other any) bool {
	var (
		m  *Loglevel
		ok bool
	)

	if m, ok = other.(*Loglevel); !ok {
		return false
	}

	return (l.Collect == m.Collect) &&
		(l.Database == m.Database) &&
		(l.DBPool == m.DBPool) &&
		(l.Monitor == m.Monitor) &&
		(l.Web == m.Web) &&
		(l.Discover == m.Discover) &&
		(l.Client == m.Client)
} // func (l *Loglevel) Equal(other any) bool

// Config represents the configuration of the app.
type Config struct {
	Global   Global
	Web      Web
	Collect  Collect
	Loglevel Loglevel
}

// Equal returns true if other is a Config value with equal values.
func (cfg *Config) Equal(other any) bool {
	// var (
	// 	c2 *Config
	// 	ok bool
	// )

	// if c2, ok = other.(*Config); !ok {
	// 	return false
	// }

	switch c2 := other.(type) {
	case *Config:
		return cfg.Global.Equal(&c2.Global) &&
			cfg.Web.Equal(&c2.Web) &&
			cfg.Collect.Equal(&c2.Collect) &&
			cfg.Loglevel.Equal(&c2.Loglevel)
	case Config:
		return cfg.Global.Equal(&c2.Global) &&
			cfg.Web.Equal(&c2.Web) &&
			cfg.Collect.Equal(&c2.Collect) &&
			cfg.Loglevel.Equal(&c2.Loglevel)
	default:
		return false
	}
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

	if cfg.Loglevel.Collect == "" {
		cfg.Loglevel.Collect = "DEBUG"
	}

	if cfg.Loglevel.Database == "" {
		cfg.Loglevel.Database = "DEBUG"
	}

	if cfg.Loglevel.DBPool == "" {
		cfg.Loglevel.DBPool = "DEBUG"
	}

	if cfg.Loglevel.Monitor == "" {
		cfg.Loglevel.Monitor = "DEBUG"
	}

	if cfg.Loglevel.Web == "" {
		cfg.Loglevel.Web = "DEBUG"
	}

	if cfg.Loglevel.Discover == "" {
		cfg.Loglevel.Discover = "DEBUG"
	}

	if cfg.Loglevel.Client == "" {
		cfg.Loglevel.Client = "DEBUG"
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
