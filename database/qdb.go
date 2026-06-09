// /home/krylon/go/src/github.com/blicero/hertz/database/qdb.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-09 13:04:13 krylon>

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
	query.HostGetAll: `
SELECT
    id,
    name,
    last_contact
FROM host
`,
	query.HostUpdateLastContact: "UPDATE host SET last_contact = ? WHERE id = ?",
	query.HostDelete:            "DELETE FROM host WHERE id = ?",
	query.RecordAdd: `
INSERT INTO record (host_id, timestamp, freq) VALUES (?, ?, ?)
ON CONFLICT (host_id, timestamp) DO NOTHING
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
WHERE timestamp BETWEEN ? AND ?
ORDER BY timestamp DESC
`,
	query.RecordGetByHostPeriod: `
SELECT
    id,
    timestamp,
    freq
FROM record
WHERE host_id = ? AND timestamp BETWEEN ? AND ?
ORDER BY timestamp DESC
`,
}
