// /home/krylon/go/src/github.com/blicero/hertz/discover/discover.go
// -*- mode: go; coding: utf-8; -*-
// Created on 03. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-09 10:44:01 krylon>

// Package discover implements peer discovery for a networked environment.
package discover

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blicero/hertz/client"
	"github.com/blicero/hertz/common"
	"github.com/blicero/hertz/common/control"
	"github.com/blicero/hertz/logdomain"
	"github.com/schollz/peerdiscovery"
	pd "github.com/schollz/peerdiscovery"
)

// Explorer looks for peers on the network.
type Explorer struct {
	log       *log.Logger
	mode      string
	server    string
	peer      *pd.PeerDiscovery
	active    atomic.Bool
	xinterval time.Duration
	lock      sync.RWMutex
	peers     map[string]string
	client    *client.Client
	cmdQ      chan control.Message
}

// Create creates a new Explorer.
func Create(mode string, xinterval time.Duration) (*Explorer, error) {
	var (
		err error
		opt pd.Settings
		xp  = &Explorer{
			mode:      mode,
			peers:     make(map[string]string),
			cmdQ:      make(chan control.Message, 2),
			xinterval: xinterval,
		}
	)

	opt.Limit = -1
	opt.TimeLimit = -1
	opt.Payload = []byte(mode)
	// opt.IPVersion = pd.IPv6
	opt.Notify = xp.handleNewPeer
	opt.NotifyLost = xp.handleLostPeer
	opt.Port = strconv.FormatInt(common.WebPort, 10)

	if xp.log, err = common.GetLogger(logdomain.Discover); err != nil {
		return nil, err
	} else if xp.peer, err = pd.NewPeerDiscovery(opt); err != nil {
		xp.log.Printf("[ERROR] Cannot initialize peer discovery: %s\n",
			err.Error())
		return nil, err
	}

	xp.active.Store(true)

	return xp, nil
} // func Create(mode string) (*Explorer, error)

// Shutdown stops the Explorer's activity.
func (xp *Explorer) Shutdown() {
	xp.lock.Lock()
	xp.peer.Shutdown()
	xp.lock.Unlock()
	xp.active.Store(false)
}

func (xp *Explorer) handleNewPeer(info peerdiscovery.Discovered) {
	xp.lock.Lock()
	defer xp.lock.Unlock()

	if _, ok := xp.peers[info.Address]; !ok {
		var (
			err error
			pl  = string(info.Payload)
		)
		xp.log.Printf("[DEBUG] Discovered new peer %s -- %s\n",
			info.Address,
			pl)
		xp.peers[info.Address] = pl

		if xp.mode != "server" && pl == "server" && xp.server == "" {
			xp.server = pl
			var srvAddr = fmt.Sprintf("http://%s:%d",
				info.Address,
				common.WebPort)
			if xp.client, err = client.New(srvAddr, xp.xinterval, xp.cmdQ); err != nil {
				xp.log.Printf("[ERROR] Failed to create Client: %s\n",
					err.Error())
			} else {
				xp.client.Start()
				xp.log.Printf("[DEBUG] Starting Client to send data to %s\n",
					xp.server)
			}
		}
	}
} // func (xp *Explorer) handleNewPeer(info peerdiscovery.Discovered)

func (xp *Explorer) handleLostPeer(lost pd.LostPeer) {
	xp.log.Printf("[DEBUG] Peer %s disappeared -- %s\n",
		lost.Address,
		lost.LastPayload)

	xp.lock.Lock()
	defer xp.lock.Unlock()

	delete(xp.peers, lost.Address)

	if xp.mode != "server" && xp.server == lost.Address && xp.client != nil {
		xp.log.Printf("[DEBUG] Lost touch with server %s, stopping Client\n",
			xp.server)
		xp.server = ""
		xp.cmdQ <- control.Message{
			Cmd:     control.Stop,
			Payload: "He dead, Jim",
		}
	}
} // func (xp *Explorer) handleLostPeer(lost pd.LostPeer)
