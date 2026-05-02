// Copyright 2026 Tim Jespers
// SPDX-License-Identifier: Apache-2.0

package status

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjespers/dofs/internal/config"
	"github.com/tjespers/dofs/internal/db"
	"github.com/tjespers/dofs/internal/linker"
	"github.com/tjespers/dofs/internal/model"
)

type Reporter struct {
	Cfg   *config.Config
	Store *db.Store
}

func (r *Reporter) Report() (*model.StatusReport, error) {
	l := &linker.Linker{Cfg: r.Cfg, Store: r.Store}
	results, err := l.CheckAll(nil)
	if err != nil {
		return nil, fmt.Errorf("check links: %w", err)
	}

	report := &model.StatusReport{
		Root:  r.Cfg.Root,
		Links: results,
	}

	// Compute summary
	for _, lr := range results {
		report.Summary.TotalLinks++
		switch lr.Status {
		case model.LinkOK:
			report.Summary.OKCount++
		case model.LinkMissing:
			report.Summary.MissingCount++
		case model.LinkBroken:
			report.Summary.BrokenCount++
		case model.LinkConflict:
			report.Summary.ConflictCount++
		case model.LinkWrongTarget:
			report.Summary.WrongCount++
		}
	}

	// Find untracked items in linkable categories
	report.UntrackedItems = r.findUntrackedItems(results)

	// Find untracked symlinks in ~/ pointing into DOFS root
	report.UntrackedLinks = r.findUntrackedLinks(results)

	// Last backup per category
	report.LastBackups, _ = r.Store.LastBackupPerCategory()
	if report.LastBackups == nil {
		report.LastBackups = make(map[model.Category]*model.BackupRecord)
	}

	return report, nil
}

func (r *Reporter) findUntrackedItems(results []model.LinkCheckResult) []string {
	tracked := make(map[string]bool)
	for _, lr := range results {
		tracked[lr.Link.SourcePath] = true
	}

	var untracked []string
	for _, cat := range model.LinkableCategories {
		dir := filepath.Join(r.Cfg.Root, string(cat))
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			path := filepath.Join(dir, e.Name())
			if !tracked[path] {
				rel, _ := filepath.Rel(r.Cfg.Root, path)
				untracked = append(untracked, rel)
			}
		}
	}
	return untracked
}

func (r *Reporter) findUntrackedLinks(results []model.LinkCheckResult) []string {
	tracked := make(map[string]bool)
	for _, lr := range results {
		tracked[lr.Link.TargetPath] = true
	}

	var untracked []string
	entries, err := os.ReadDir(r.Cfg.HomeDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		fullPath := filepath.Join(r.Cfg.HomeDir, e.Name())
		if tracked[fullPath] {
			continue
		}
		info, err := os.Lstat(fullPath)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			continue
		}
		target, err := os.Readlink(fullPath)
		if err != nil {
			continue
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(fullPath), target)
		}
		target, _ = filepath.Abs(target)
		if strings.HasPrefix(target, r.Cfg.Root) {
			untracked = append(untracked, e.Name())
		}
	}
	return untracked
}
