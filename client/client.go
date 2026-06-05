// /home/krylon/go/src/github.com/blicero/hertz/client/client.go
// -*- mode: go; coding: utf-8; -*-
// Created on 05. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-05 15:39:55 krylon>

// Package client handles communication with a Server.
package client

import (
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blicero/hertz/common"
	"github.com/blicero/hertz/common/control"
	"github.com/blicero/hertz/database"
	"github.com/blicero/hertz/logdomain"
)

// Client wraps the state for communicating with a Server.
type Client struct {
	log    *log.Logger
	addr   string
	active atomic.Bool
	client *http.Client
	wg     *sync.WaitGroup
	db     *database.Database
	cmdQ   <-chan control.Message
}

func New(addr string, cmdQ <-chan control.Message) (*Client, error) {
	var (
		err error
		c   = &Client{
			addr: addr,
			cmdQ: cmdQ,
			wg:   new(sync.WaitGroup),
			client: &http.Client{
				Timeout: time.Millisecond * 2500,
			},
		}
	)

	if c.log, err = common.GetLogger(logdomain.Client); err != nil {
		return nil, err
	} else if c.db, err = database.Open(common.DbPath); err != nil {
		c.log.Printf("[CRITICAL] Failed to open database: %s\n",
			err.Error())
		return nil, err
	}

	return c, nil
} // func New(addr string, cmdQ <-chan control.Message) (*Client, error)

// Start starts the client main loop.
func (c *Client) Start() {
	c.active.Store(true)

} // func (c *Client) Start()

// Stop shuts down the client.
func (c *Client) Stop() {
	c.active.Store(false)
}
