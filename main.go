// /home/krylon/go/src/github.com/blicero/hertz/main.go
// -*- mode: go; coding: utf-8; -*-
// Created on 30. 05. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-05-30 11:44:22 krylon>

package main

import (
	"fmt"

	"github.com/blicero/hertz/common"
)

func main() {
	fmt.Printf("%s %s\n",
		common.AppName,
		common.Version)
	fmt.Println("\nNothing to see here (yet), move along...")
}
