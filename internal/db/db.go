// Copyright 2026 Tim Jespers
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tjespers/dofs/internal/model"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(wal)&_pragma=foreign_keys(on)")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER NOT NULL,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)
	if err != nil {
		return err
	}

	var current int
	row := s.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version")
	if err := row.Scan(&current); err != nil {
		return err
	}

	for i := current; i < len(migrations); i++ {
		if _, err := s.db.Exec(migrations[i]); err != nil {
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
		if _, err := s.db.Exec("INSERT INTO schema_version (version) VALUES (?)", i+1); err != nil {
			return err
		}
	}
	return nil
}

// --- Managed Links ---

func (s *Store) InsertLink(link *model.ManagedLink) error {
	_, err := s.db.Exec(
		`INSERT INTO managed_links (category, source_name, source_path, target_path, last_status)
		 VALUES (?, ?, ?, ?, ?)`,
		link.Category, link.SourceName, link.SourcePath, link.TargetPath, link.LastStatus,
	)
	return err
}

func (s *Store) ListLinks(category *model.Category) ([]model.ManagedLink, error) {
	query := `SELECT id, category, source_name, source_path, target_path, last_status,
	          created_at, updated_at, last_checked_at FROM managed_links`
	var args []any
	if category != nil {
		query += " WHERE category = ?"
		args = append(args, string(*category))
	}
	query += " ORDER BY category, source_name"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.ManagedLink
	for rows.Next() {
		var l model.ManagedLink
		var createdAt, updatedAt string
		var lastCheckedAt sql.NullString
		if err := rows.Scan(&l.ID, &l.Category, &l.SourceName, &l.SourcePath,
			&l.TargetPath, &l.LastStatus, &createdAt, &updatedAt, &lastCheckedAt); err != nil {
			return nil, err
		}
		l.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		l.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		if lastCheckedAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", lastCheckedAt.String)
			l.LastCheckedAt = &t
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func (s *Store) GetLinkByTarget(targetPath string) (*model.ManagedLink, error) {
	var l model.ManagedLink
	var createdAt, updatedAt string
	var lastCheckedAt sql.NullString
	err := s.db.QueryRow(
		`SELECT id, category, source_name, source_path, target_path, last_status,
		 created_at, updated_at, last_checked_at FROM managed_links WHERE target_path = ?`,
		targetPath,
	).Scan(&l.ID, &l.Category, &l.SourceName, &l.SourcePath,
		&l.TargetPath, &l.LastStatus, &createdAt, &updatedAt, &lastCheckedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	l.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	l.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	if lastCheckedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", lastCheckedAt.String)
		l.LastCheckedAt = &t
	}
	return &l, nil
}

func (s *Store) UpdateLinkStatus(id int64, status model.LinkStatus) error {
	_, err := s.db.Exec(
		`UPDATE managed_links SET last_status = ?, last_checked_at = datetime('now'),
		 updated_at = datetime('now') WHERE id = ?`,
		status, id,
	)
	return err
}

func (s *Store) DeleteLink(id int64) error {
	_, err := s.db.Exec("DELETE FROM managed_links WHERE id = ?", id)
	return err
}

// --- Scan ---

func (s *Store) InsertScanResult(path string, disposition model.ScanDisposition, category *model.Category) error {
	var cat *string
	if category != nil {
		c := string(*category)
		cat = &c
	}
	_, err := s.db.Exec(
		`INSERT INTO scan_history (found_path, disposition, adopted_category) VALUES (?, ?, ?)`,
		path, disposition, cat,
	)
	return err
}

func (s *Store) ListIgnorePatterns() ([]string, error) {
	rows, err := s.db.Query("SELECT pattern FROM scan_ignores ORDER BY pattern")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var patterns []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		patterns = append(patterns, p)
	}
	return patterns, rows.Err()
}

func (s *Store) AddIgnorePattern(pattern string) error {
	_, err := s.db.Exec("INSERT OR IGNORE INTO scan_ignores (pattern) VALUES (?)", pattern)
	return err
}

// --- Backup ---

func (s *Store) InsertBackup(record *model.BackupRecord) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO backup_history (category, archive_path, status) VALUES (?, ?, ?)`,
		record.Category, record.ArchivePath, "in_progress",
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) CompleteBackup(id int64, sizeBytes int64) error {
	_, err := s.db.Exec(
		`UPDATE backup_history SET status = 'completed', size_bytes = ?,
		 completed_at = datetime('now') WHERE id = ?`,
		sizeBytes, id,
	)
	return err
}

func (s *Store) FailBackup(id int64) error {
	_, err := s.db.Exec(
		`UPDATE backup_history SET status = 'failed', completed_at = datetime('now') WHERE id = ?`,
		id,
	)
	return err
}

func (s *Store) LastBackupPerCategory() (map[model.Category]*model.BackupRecord, error) {
	rows, err := s.db.Query(
		`SELECT id, category, archive_path, COALESCE(size_bytes, 0), started_at, completed_at, status
		 FROM backup_history
		 WHERE status = 'completed'
		 AND id IN (SELECT MAX(id) FROM backup_history WHERE status = 'completed' GROUP BY category)`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[model.Category]*model.BackupRecord)
	for rows.Next() {
		var r model.BackupRecord
		var startedAt string
		var completedAt sql.NullString
		if err := rows.Scan(&r.ID, &r.Category, &r.ArchivePath, &r.SizeBytes,
			&startedAt, &completedAt, &r.Status); err != nil {
			return nil, err
		}
		r.StartedAt, _ = time.Parse("2006-01-02 15:04:05", startedAt)
		if completedAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", completedAt.String)
			r.CompletedAt = &t
		}
		result[r.Category] = &r
	}
	return result, rows.Err()
}
