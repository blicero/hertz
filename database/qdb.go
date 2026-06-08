// /home/krylon/go/src/github.com/blicero/hertz/database/qdb.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-08 12:18:00 krylon>

package database

import "github.com/blicero/hertz/database/query"

var qdb = map[query.ID]string{
	query.HostAdd: "INSERT INTO host (name) VALUES (?) RETURNING id",
	query.HostGetByName: `
SELECT
    id,
    last_contact
FROM host
WHERE name = ?
`,
	query.HostGetByID: `
SELECT
    name,
    last_contact
FROM host
WHERE id = ?
`,
	query.HostGetRecent: `
SELECT
    id,
    name,
    last_contact
FROM host
WHERE last_contact >= ?
`,
	query.HostUpdateLastContact: "UPDATE host SET last_contact = ? WHERE id = ?",
	query.HostDelete:            "DELETE FROM host WHERE id = ?",
	query.RecordAdd: `
INSERT INTO record (host_id, timestamp, freq) VALUES (?, ?, ?)
RETURNING id
`,
	query.RecordGetByHost: `
SELECT
    id,
    timestamp,
    freq
FROM record
WHERE host_id = ?
ORDER BY timestamp DESC
LIMIT ?
`,
	query.RecordGetByPeriod: `
SELECT
    id,
    host_id,
    timestamp,
    freq
FROM record
ORDER BY timestamp DESC
WHERE timestamp BETWEEN ? AND ?
`,
	query.RecordGetByHostPeriod: `
SELECT
    id,
    timestamp,
    freq
FROM record
ORDER BY timestamp DESC
WHERE host_id = ? AND timestamp BETWEEN ? AND ?
`,
}
