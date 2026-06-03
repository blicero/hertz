// /home/krylon/go/src/github.com/blicero/hertz/discover/discover.go
// -*- mode: go; coding: utf-8; -*-
// Created on 03. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-03 13:39:01 krylon>

// Package discover implements peer discovery for a networked environment.
package discover

import (
	"log"
	"sync"
	"sync/atomic"

	"github.com/blicero/hertz/common"
	"github.com/blicero/hertz/logdomain"
	"github.com/schollz/peerdiscovery"
	pd "github.com/schollz/peerdiscovery"
)

// Explorer looks for peers on the network.
type Explorer struct {
	log    *log.Logger
	peer   *pd.PeerDiscovery
	active atomic.Bool
	lock   sync.RWMutex
	peers  map[string]string
}

// Create creates a new Explorer.
func Create(mode string) (*Explorer, error) {
	var (
		err error
		opt pd.Settings
		xp  = &Explorer{peers: make(map[string]string)}
	)

	opt.Limit = -1
	opt.Payload = []byte(mode)
	opt.IPVersion = pd.IPv6
	opt.Notify = xp.handleNewPeer
	opt.NotifyLost = xp.handleLostPeer

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
	xp.log.Printf("[DEBUG] Discovered new peer %s -- %s\n",
		info.Address,
		info.Payload)
	xp.lock.Lock()
	defer xp.lock.Unlock()

	if _, ok := xp.peers[info.Address]; !ok {
		xp.peers[info.Address] = string(info.Payload)
	}
} // func (xp *Explorer) handleNewPeer(info peerdiscovery.Discovered)

func (xp *Explorer) handleLostPeer(lost pd.LostPeer) {
	xp.log.Printf("[DEBUG] Discovered new peer %s -- %s\n",
		lost.Address,
		lost.LastPayload)
	xp.lock.Lock()
	delete(xp.peers, lost.Address)
	xp.lock.Unlock()
} // func (xp *Explorer) handleLostPeer(lost pd.LostPeer)
