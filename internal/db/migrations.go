// Copyright 2026 Tim Jespers
// SPDX-License-Identifier: Apache-2.0

package db

var migrations = []string{
	// v1: initial schema
	`CREATE TABLE IF NOT EXISTS managed_links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		category TEXT NOT NULL,
		source_name TEXT NOT NULL,
		source_path TEXT NOT NULL,
		target_path TEXT NOT NULL UNIQUE,
		last_status TEXT NOT NULL DEFAULT 'unknown',
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		updated_at TEXT NOT NULL DEFAULT (datetime('now')),
		last_checked_at TEXT,
		UNIQUE(category, source_name)
	);

	CREATE TABLE IF NOT EXISTS scan_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		found_path TEXT NOT NULL,
		disposition TEXT NOT NULL DEFAULT 'pending',
		adopted_category TEXT,
		scanned_at TEXT NOT NULL DEFAULT (datetime('now')),
		resolved_at TEXT
	);

	CREATE TABLE IF NOT EXISTS scan_ignores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		pattern TEXT NOT NULL UNIQUE,
		created_at TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS backup_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		category TEXT NOT NULL,
		archive_path TEXT NOT NULL,
		size_bytes INTEGER,
		started_at TEXT NOT NULL DEFAULT (datetime('now')),
		completed_at TEXT,
		status TEXT NOT NULL DEFAULT 'in_progress'
	);

	CREATE INDEX IF NOT EXISTS idx_managed_links_category ON managed_links(category);
	CREATE INDEX IF NOT EXISTS idx_managed_links_status ON managed_links(last_status);
	CREATE INDEX IF NOT EXISTS idx_scan_history_disposition ON scan_history(disposition);`,
}
