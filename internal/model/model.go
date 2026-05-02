// Copyright 2026 Tim Jespers
// SPDX-License-Identifier: Apache-2.0

package model

import "time"

type Category string

const (
	CategoryHome      Category = "home"
	CategorySecrets   Category = "secrets"
	CategoryCode      Category = "code"
	CategoryData      Category = "data"
	CategoryDesktop   Category = "desktop"
	CategoryDocuments Category = "documents"
	CategoryDownloads Category = "downloads"
	CategoryMusic     Category = "music"
	CategoryPictures  Category = "pictures"
)

var LinkableCategories = []Category{CategoryHome, CategorySecrets}

var AllCategories = []Category{
	CategoryHome, CategorySecrets, CategoryCode,
	CategoryData, CategoryDesktop, CategoryDocuments,
	CategoryDownloads, CategoryMusic, CategoryPictures,
}

type LinkStatus string

const (
	LinkOK          LinkStatus = "ok"
	LinkMissing     LinkStatus = "missing"
	LinkBroken      LinkStatus = "broken"
	LinkConflict    LinkStatus = "conflict"
	LinkWrongTarget LinkStatus = "wrong_target"
	LinkUnknown     LinkStatus = "unknown"
)

type ManagedLink struct {
	ID            int64
	Category      Category
	SourceName    string
	SourcePath    string
	TargetPath    string
	LastStatus    LinkStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastCheckedAt *time.Time
}

type LinkCheckResult struct {
	Link         ManagedLink
	Status       LinkStatus
	ActualTarget string
	Error        error
}

type ScanCandidate struct {
	Name  string
	Path  string
	IsDir bool
	Size  int64
}

type ScanDisposition string

const (
	ScanAdopted ScanDisposition = "adopted"
	ScanIgnored ScanDisposition = "ignored"
	ScanPending ScanDisposition = "pending"
)

type BackupRecord struct {
	ID          int64
	Category    Category
	ArchivePath string
	SizeBytes   int64
	StartedAt   time.Time
	CompletedAt *time.Time
	Status      string
}

type StatusReport struct {
	Root           string
	Links          []LinkCheckResult
	UntrackedItems []string
	UntrackedLinks []string
	LastBackups    map[Category]*BackupRecord
	Summary        StatusSummary
}

type StatusSummary struct {
	TotalLinks    int
	OKCount       int
	MissingCount  int
	BrokenCount   int
	ConflictCount int
	WrongCount    int
}
