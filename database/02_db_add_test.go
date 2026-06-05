// /home/krylon/go/src/github.com/blicero/hertz/database/02_db_add_test.go
// -*- mode: go; coding: utf-8; -*-
// Created on 05. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-05 20:23:27 krylon>

package database

import (
	"math/rand"
	"testing"
	"time"

	"github.com/blicero/hertz/model"
)

const (
	recCnt = 120
)

var (
	begin   time.Time
	records []*model.Record
)

func TestDBRecordAdd(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	}

	const (
		maxFreq = 3200
	)

	var (
		err       error
		baseStamp = time.Now().Add(time.Second * -7200)
	)

	begin = baseStamp
	records = make([]*model.Record, recCnt)

	for i := range recCnt {
		var rec = &model.Record{
			Timestamp: baseStamp.Add(time.Second * 60 * time.Duration(i)),
			Freq:      make([]int64, 4),
		}

		var f1, f2 int64

		f1 = rand.Int63n(maxFreq)
		f2 = rand.Int63n(maxFreq)

		rec.Freq[0] = f1
		rec.Freq[1] = f1
		rec.Freq[2] = f2
		rec.Freq[3] = f2

		if err = tdb.RecordAdd(rec); err != nil {
			t.Fatalf("Failed to add Record: %s\n",
				err.Error())
		} else if rec.ID == 0 {
			t.Errorf("Unexpected Record ID %d (should be %d)\n",
				rec.ID,
				i+1)
		}

		records[i] = rec
	}
} // func TestDBRecordAdd(t *testing.T)

func TestDBRecordGet(t *testing.T) {
	if tdb == nil {
		t.SkipNow()
	} else if records == nil {
		t.SkipNow()
	}

	var (
		err       error
		dbRecords []*model.Record
	)

	if dbRecords, err = tdb.RecordGet(begin); err != nil {
		t.Fatalf("Failed to load Records: %s\n",
			err.Error())
	} else if len(dbRecords) != recCnt {
		t.Fatalf("Unexpected number of records: %d (expected %d)",
			len(dbRecords),
			recCnt)
	}

	for i, r1 := range dbRecords {
		r2 := records[i]

		if !r1.Equal(r2) {
			t.Fatalf(`Records %d are not equal:
Original: %#v,
Fetched:  %#v
`,
				r1.ID,
				r1,
				r2)
		}
	}
} // func TestDBRecordGet(t *testing.T)
