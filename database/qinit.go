// /home/krylon/go/src/github.com/blicero/hertz/database/qinit.go
// -*- mode: go; coding: utf-8; -*-
// Created on 08. 06. 2026 by Benjamin Walkenhorst
// (c) 2026 Benjamin Walkenhorst
// Time-stamp: <2026-06-08 12:50:47 krylon>

package database

var qinit = []string{
	`
CREATE TABLE host (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    last_contact INTEGER NOT NULL DEFAULT 0
) STRICT
`,
	"CREATE INDEX host_name_idx ON host (name)",
	"CREATE INDEX host_contact_idx ON host (last_contact)",
	`
CREATE TABLE record (
    id INTEGER PRIMARY KEY,
    host_id INTEGER NOT NULL,
    timestamp INTEGER NOT NULL,
    freq TEXT NOT NULL,
    UNIQUE (host_id, timestamp),
    CHECK (json_valid(freq)),
    FOREIGN KEY (host_id) REFERENCES host (id)
      ON UPDATE RESTRICT
      ON DELETE CASCADE
) STRICT
`,
	"CREATE INDEX time_idx ON record (timestamp)",
	`
CREATE TRIGGER IF NOT EXISTS host_contact_tr
AFTER INSERT ON record
BEGIN
    UPDATE host SET last_contact = new.timestamp;
END;
`,
}
