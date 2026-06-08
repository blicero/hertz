// /home/krylon/go/src/github.com/blicero/hertz/database/host.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-08 12:28:12 krylon>

package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/blicero/hertz/database/query"
	"github.com/blicero/hertz/model"
)

// HostAdd adds a new Host to the database.
func (db *Database) HostAdd(host *model.Host) error {
	const qid query.ID = query.HostAdd
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

	var rows *sql.Rows
EXEC_QUERY:
	if rows, err = stmt.Query(host.Name); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			err = fmt.Errorf("cannot add Host %s: %w",
				host.Name,
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
			var ex = fmt.Errorf("failed to get ID for newly added Host %s: %w",
				host.Name,
				err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return ex
		}

		host.ID = id
	}

	return nil
} // func (db *Database) HostAdd(host *model.Host) error

// HostGetByName looks up a Host by its name.
func (db *Database) HostGetByName(name string) (*model.Host, error) {
	const qid query.ID = query.HostGetByName
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
	if rows, err = stmt.Query(name); err != nil {
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

	if rows.Next() {
		var (
			lastContact int64
			host        = &model.Host{Name: name}
		)

		if err = rows.Scan(&host.ID, &lastContact); err != nil {
			var ex = fmt.Errorf("failed to scan row: %w", err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		}

		host.LastContact = time.Unix(lastContact, 0)
		return host, nil
	}

	return nil, nil
} // func (db *Database) HostGetByName(name string) (*model.Host, error)

// HostGetByID looks up a Host by its ID.
func (db *Database) HostGetByID(id int64) (*model.Host, error) {
	const qid query.ID = query.HostGetByID
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
	if rows, err = stmt.Query(id); err != nil {
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

	if rows.Next() {
		var (
			lastContact int64
			host        = &model.Host{ID: id}
		)

		if err = rows.Scan(&host.Name, &lastContact); err != nil {
			var ex = fmt.Errorf("failed to scan row: %w", err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		}

		host.LastContact = time.Unix(lastContact, 0)
		return host, nil
	}

	return nil, nil
} // func (db *Database) HostGetByID(id int64) (*model.Host, error)

// HostGetRecent fetches all Hosts that submitted data since the givem timestamp.
func (db *Database) HostGetRecent(after time.Time) ([]*model.Host, error) {
	const qid query.ID = query.HostGetRecent
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
	if rows, err = stmt.Query(after.Unix()); err != nil {
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
	var hosts = make([]*model.Host, 0)

	for rows.Next() {
		var (
			lastContact int64
			host        = new(model.Host)
		)

		if err = rows.Scan(&host.ID, &host.Name, &lastContact); err != nil {
			var ex = fmt.Errorf("failed to scan row: %w", err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return nil, ex
		}

		host.LastContact = time.Unix(lastContact, 0)
		hosts = append(hosts, host)
	}

	return hosts, nil
} // func (db *Database) HostGetRecent(after time.Time) ([]*model.Host, error)

// HostUpdateLastContact updates a Host's contact timestamp.
func (db *Database) HostUpdateLastContact(host *model.Host, when time.Time) error {
	const qid query.ID = query.HostUpdateLastContact
	var (
		err, ex error
		stmt    *sql.Stmt
		res     sql.Result
		cnt     int64
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Failed to prepare query %s: %s\n",
			qid,
			err.Error())
		panic(err)
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

EXEC_QUERY:
	if res, err = stmt.Exec(when.Unix(), host.ID); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			ex = fmt.Errorf("cannot update last contact of Host %s (%d): %w",
				host.Name,
				host.ID,
				err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return ex
		}
	} else if cnt, err = res.RowsAffected(); err != nil {
		ex = fmt.Errorf("failed to get number of affected rows: %w",
			err)
		db.log.Printf("[ERROR] %s\n", ex.Error())
		return ex
	} else if cnt != 1 {
		ex = fmt.Errorf("unexpected number of affected rows for %s: %d (expected 1)",
			qid,
			cnt)
		db.log.Printf("[CRITICAL] %s\n", ex.Error())
		return ex
	}

	host.LastContact = when
	return nil
} // func (db *Database) HostUpdateLastContact(host *model.Host, when time.Time) error

// HostDelete deletes a Host from the Database.
func (db *Database) HostDelete(host *model.Host) error {
	const qid query.ID = query.HostDelete
	var (
		err, ex error
		stmt    *sql.Stmt
		res     sql.Result
		cnt     int64
	)

	if stmt, err = db.getQuery(qid); err != nil {
		db.log.Printf("[ERROR] Failed to prepare query %s: %s\n",
			qid,
			err.Error())
		panic(err)
	} else if db.tx != nil {
		stmt = db.tx.Stmt(stmt)
	}

EXEC_QUERY:
	if res, err = stmt.Exec(host.ID); err != nil {
		if worthARetry(err) {
			waitForRetry()
			goto EXEC_QUERY
		} else {
			ex = fmt.Errorf("cannot delete Host %s (%d): %w",
				host.Name,
				host.ID,
				err)
			db.log.Printf("[ERROR] %s\n", ex.Error())
			return ex
		}
	} else if cnt, err = res.RowsAffected(); err != nil {
		ex = fmt.Errorf("failed to get number of affected rows: %w",
			err)
		db.log.Printf("[ERROR] %s\n", ex.Error())
		return ex
	} else if cnt != 1 {
		ex = fmt.Errorf("unexpected number of affected rows for %s: %d (expected 1)",
			qid,
			cnt)
		db.log.Printf("[CRITICAL] %s\n", ex.Error())
		return ex
	}

	return nil
} // func (db *Database) HostDelete(host *model.Host) error
