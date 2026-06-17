// /home/krylon/go/src/hertz/web/tmpl_data.go
// -*- mode: go; coding: utf-8; -*-
// Created on 06. 05. 2020 by Benjamin Walkenhorst
// (c) 2020 Benjamin Walkenhorst
// Time-stamp: <2026-06-17 12:04:20 krylon>
//
// This file contains data structures to be passed to HTML templates.

package web

import "github.com/blicero/hertz/model"

type tmplDataBase struct {
	Title    string
	Debug    bool
	URL      string
	Messages []string
	Hosts    []*model.Host
}

type tmplDataIndex struct {
	tmplDataBase
}

type tmplDataHosts struct {
	tmplDataBase
}

type tmplDataSingleHost struct {
	tmplDataBase
	Host        *model.Host
	Records     []*model.Record
	Histogram   map[int64]int64
	HistPercent map[int64]float64
	Stride      int
}
