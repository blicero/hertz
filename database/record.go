// /home/krylon/go/src/github.com/blicero/hertz/database/record.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-08 12:51:10 krylon>

package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/blicero/hertz/database/query"
	"github.com/blicero/hertz/model"
)

// RecordAdd stores a Record in the Database.
func (db *Database) RecordAdd(rec *model.Record) error {
	const qid query.ID = query.RecordAdd
	var (
		err  error
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Failed to prepare query %s: %s\n",
			qid,
			err.Error())
		panic(err)
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var (
		rows *sql.Rows
		freq []byte
	)

	if freq, err = json.Marshal(rec.Freq); err != nil {
		db.log.Printf("[CANTHAPPEN] Cannot serialize frequency data: %s\n",
			err.Error())
		return err
	}
EXEC_QUERY:
	if rows, err = stmt.Query(rec.HostID, rec.Timestamp, string(freq)); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			err = fmt.Errorf("cannot add Record for Host %d: %w",
				rec.HostID,
				err)
			db.log.Printf("[ERROR] %s\n", err.Error())
			return err
		}
	} else {
		var id int64

		defer rows.Close() // nolint: errcheck

		if !rows.Next() {
			// CANTHAPPEN
			db.log.Printf("[ERROR] Query %s did not return a value\n",
				qid)
			return fmt.Errorf("query %s did not return a value", qid)
		} else if err = rows.Scan(&id); err != nil {
			var ex = fmt.Errorf("failed to get ID for newly added Record: %w",
				err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return ex
		}

		rec.ID = id
	}

	return nil
} // func (db *Database) RecordAdd(record *model.Record) error

// RecordGetByHost loads up to <limit> Records from the given Host, ordered
// by their timestamps in descending order.
func (db *Database) RecordGetByHost(host *model.Host, limit int64) ([]*model.Record, error) {
	const qid query.ID = query.RecordGetByHost
	var (
		err  error
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(host.ID, limit); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			db.log.Printf("[ERROR] Query %s failed: %s\n",
				qid,
				err.Error())
			return nil, err
		}
	}

	defer rows.Close() // nolint: errcheck,gosec
	var records = make([]*model.Record, 0)

	for rows.Next() {
		var (
			tstamp int64
			freq   string
			rec    = &model.Record{HostID: host.ID}
		)

		if err = rows.Scan(&rec.ID, &tstamp, &freq); err != nil {
			var ex = fmt.Errorf("failed to scan row: %w", err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		} else if err = json.Unmarshal([]byte(freq), &rec.Freq); err != nil {
			var ex = fmt.Errorf("failed to parse frequency from JSON: %s\n%s",
				err.Error(),
				freq)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		}

		rec.Timestamp = time.Unix(tstamp, 0)
		records = append(records, rec)
	}

	return records, nil
} // func (db *Database) RecordGetByHost(host *model.Host, limit int64) ([]*model.Record, error)

// RecordGetByPeriod loads all Records from the given timespan in reverse
// chronological order
func (db *Database) RecordGetByPeriod(begin, end time.Time) ([]*model.Record, error) {
	const qid query.ID = query.RecordGetByPeriod
	var (
		err  error
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(begin.Unix(), end.Unix()); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			db.log.Printf("[ERROR] Query %s failed: %s\n",
				qid,
				err.Error())
			return nil, err
		}
	}

	defer rows.Close() // nolint: errcheck,gosec
	var records = make([]*model.Record, 0)

	for rows.Next() {
		var (
			tstamp int64
			freq   string
			rec    = new(model.Record)
		)

		if err = rows.Scan(&rec.ID, &rec.HostID, &tstamp, &freq); err != nil {
			var ex = fmt.Errorf("failed to scan row: %w", err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		} else if err = json.Unmarshal([]byte(freq), &rec.Freq); err != nil {
			var ex = fmt.Errorf("failed to parse frequency from JSON: %s\n%s",
				err.Error(),
				freq)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		}

		rec.Timestamp = time.Unix(tstamp, 0)
		records = append(records, rec)
	}

	return records, nil
} // func (db *Database) RecordGetByPeriod(begin, end time.Time) ([]*model.Record, error)

// RecordGetByHostPeriod loads all Records from the given Host and timespan.
func (db *Database) RecordGetByHostPeriod(host *model.Host, begin, end time.Time) ([]*model.Record, error) {
	const qid query.ID = query.RecordGetByHostPeriod
	var (
		err  error
		stmt *sql.Stmt
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Cannot prepare query %s: %s\n",
			qid,
			err.Error())
		return nil, err
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

	var rows *sql.Rows

EXEC_QUERY:
	if rows, err = stmt.Query(host.ID, begin.Unix(), end.Unix()); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			db.log.Printf("[ERROR] Query %s failed: %s\n",
				qid,
				err.Error())
			return nil, err
		}
	}

	defer rows.Close() // nolint: errcheck,gosec
	var records = make([]*model.Record, 0)

	for rows.Next() {
		var (
			tstamp int64
			freq   string
			rec    = &model.Record{HostID: host.ID}
		)

		if err = rows.Scan(&rec.ID, &tstamp, &freq); err != nil {
			var ex = fmt.Errorf("failed to scan row: %w", err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		} else if err = json.Unmarshal([]byte(freq), &rec.Freq); err != nil {
			var ex = fmt.Errorf("failed to parse frequency from JSON: %s\n%s",
				err.Error(),
				freq)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		}

		rec.Timestamp = time.Unix(tstamp, 0)
		records = append(records, rec)
	}

	return records, nil
} // func (db *Database) RecordGetByHostPeriod(host *model.Host, begin, end time.Time) ([]*model.Record, error)
