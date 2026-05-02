// Copyright 2026 Tim Jespers
// SPDX-License-Identifier: Apache-2.0

package linker

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tjespers/dofs/internal/config"
	"github.com/tjespers/dofs/internal/db"
	"github.com/tjespers/dofs/internal/model"
)

type Linker struct {
	Cfg   *config.Config
	Store *db.Store
}

func (l *Linker) CheckAll(category *model.Category) ([]model.LinkCheckResult, error) {
	links, err := l.Store.ListLinks(category)
	if err != nil {
		return nil, fmt.Errorf("list links: %w", err)
	}
	var results []model.LinkCheckResult
	for _, link := range links {
		r := l.CheckOne(link)
		_ = l.Store.UpdateLinkStatus(link.ID, r.Status)
		results = append(results, r)
	}
	return results, nil
}

func (l *Linker) CheckOne(link model.ManagedLink) model.LinkCheckResult {
	r := model.LinkCheckResult{Link: link}

	info, err := os.Lstat(link.TargetPath)
	if os.IsNotExist(err) {
		r.Status = model.LinkMissing
		return r
	}
	if err != nil {
		r.Status = model.LinkUnknown
		r.Error = err
		return r
	}

	if info.Mode()&os.ModeSymlink == 0 {
		r.Status = model.LinkConflict
		return r
	}

	actual, err := os.Readlink(link.TargetPath)
	if err != nil {
		r.Status = model.LinkUnknown
		r.Error = err
		return r
	}
	r.ActualTarget = actual

	// Check if the symlink target actually exists
	if _, err := os.Stat(link.TargetPath); os.IsNotExist(err) {
		r.Status = model.LinkBroken
		return r
	}

	// Resolve both paths to compare properly (handles relative symlinks)
	resolvedActual := actual
	if !filepath.IsAbs(actual) {
		resolvedActual = filepath.Join(filepath.Dir(link.TargetPath), actual)
	}
	var err1, err2 error
	resolvedActual, err1 = filepath.Abs(resolvedActual)
	resolvedExpected, err2 := filepath.Abs(link.SourcePath)
	if err1 != nil || err2 != nil {
		// Fall back to direct comparison
		if actual == link.SourcePath {
			r.Status = model.LinkOK
		} else {
			r.Status = model.LinkWrongTarget
		}
		return r
	}

	if resolvedActual == resolvedExpected {
		r.Status = model.LinkOK
	} else {
		r.Status = model.LinkWrongTarget
	}
	return r
}

func (l *Linker) CreateMissing(category *model.Category, force bool) (int, error) {
	links, err := l.Store.ListLinks(category)
	if err != nil {
		return 0, err
	}
	created := 0
	for _, link := range links {
		r := l.CheckOne(link)
		if r.Status != model.LinkMissing {
			continue
		}
		if _, err := os.Stat(link.SourcePath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "  skip %s: source does not exist (%s)\n", link.SourceName, link.SourcePath)
			continue
		}
		if err := l.createSymlink(link.SourcePath, link.TargetPath, force); err != nil {
			fmt.Fprintf(os.Stderr, "  error %s: %v\n", link.SourceName, err)
			continue
		}
		_ = l.Store.UpdateLinkStatus(link.ID, model.LinkOK)
		fmt.Printf("  created %s -> %s\n", link.TargetPath, link.SourcePath)
		created++
	}
	return created, nil
}

func (l *Linker) RepairBroken(category *model.Category) (int, error) {
	links, err := l.Store.ListLinks(category)
	if err != nil {
		return 0, err
	}
	repaired := 0
	for _, link := range links {
		r := l.CheckOne(link)
		if r.Status != model.LinkBroken && r.Status != model.LinkWrongTarget {
			continue
		}
		if _, err := os.Stat(link.SourcePath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "  skip %s: source does not exist (%s)\n", link.SourceName, link.SourcePath)
			continue
		}
		if l.Cfg.DryRun {
			fmt.Printf("  [dry-run] would repair %s -> %s\n", link.TargetPath, link.SourcePath)
			repaired++
			continue
		}
		if err := os.Remove(link.TargetPath); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "  error removing %s: %v\n", link.TargetPath, err)
			continue
		}
		if err := os.Symlink(link.SourcePath, link.TargetPath); err != nil {
			fmt.Fprintf(os.Stderr, "  error creating symlink %s: %v\n", link.TargetPath, err)
			continue
		}
		_ = l.Store.UpdateLinkStatus(link.ID, model.LinkOK)
		fmt.Printf("  repaired %s -> %s\n", link.TargetPath, link.SourcePath)
		repaired++
	}
	return repaired, nil
}

func (l *Linker) Add(category model.Category, sourceName, targetPath string) error {
	sourcePath := filepath.Join(l.Cfg.Root, string(category), sourceName)
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source does not exist: %s", sourcePath)
	}

	if targetPath == "" {
		targetPath = filepath.Join(l.Cfg.HomeDir, sourceName)
	}

	link := &model.ManagedLink{
		Category:   category,
		SourceName: sourceName,
		SourcePath: sourcePath,
		TargetPath: targetPath,
		LastStatus: model.LinkUnknown,
	}
	return l.Store.InsertLink(link)
}

func (l *Linker) Remove(id int64) error {
	return l.Store.DeleteLink(id)
}

func (l *Linker) createSymlink(source, target string, force bool) error {
	if l.Cfg.DryRun {
		fmt.Printf("  [dry-run] would create symlink: %s -> %s\n", target, source)
		return nil
	}

	info, err := os.Lstat(target)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			existing, _ := os.Readlink(target)
			if existing == source {
				return nil // already correct
			}
			os.Remove(target)
		} else {
			if !force {
				return fmt.Errorf("conflict: %s exists and is not a symlink (use --force)", target)
			}
			backupPath := target + ".dofs-backup." + time.Now().Format("20060102-150405")
			if err := os.Rename(target, backupPath); err != nil {
				return fmt.Errorf("backup %s: %w", target, err)
			}
			fmt.Printf("  backed up existing file to %s\n", backupPath)
		}
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	return os.Symlink(source, target)
}

// ImportExisting scans homeDir for symlinks pointing into the DOFS root and imports them.
func (l *Linker) ImportExisting() (int, error) {
	entries, err := os.ReadDir(l.Cfg.HomeDir)
	if err != nil {
		return 0, fmt.Errorf("read home dir: %w", err)
	}

	count := 0
	for _, entry := range entries {
		fullPath := filepath.Join(l.Cfg.HomeDir, entry.Name())
		info, err := os.Lstat(fullPath)
		if err != nil || info.Mode()&os.ModeSymlink == 0 {
			continue
		}

		target, err := os.Readlink(fullPath)
		if err != nil {
			continue
		}

		// Make target absolute if relative
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(fullPath), target)
		}

		// Resolve to clean absolute path
		target, err = filepath.Abs(target)
		if err != nil {
			continue
		}

		// Check if target is within DOFS root
		rel, err := filepath.Rel(l.Cfg.Root, target)
		if err != nil || len(rel) > 1 && rel[:2] == ".." {
			continue
		}

		// Determine category from path
		relParts := splitPath(rel)
		if len(relParts) < 2 {
			continue
		}

		category := model.Category(relParts[0])
		sourceName := filepath.Join(relParts[1:]...)

		// Check if already tracked
		existing, err := l.Store.GetLinkByTarget(fullPath)
		if err != nil {
			continue
		}
		if existing != nil {
			continue
		}

		link := &model.ManagedLink{
			Category:   category,
			SourceName: sourceName,
			SourcePath: target,
			TargetPath: fullPath,
			LastStatus: model.LinkOK,
		}
		if err := l.Store.InsertLink(link); err != nil {
			if l.Cfg.Verbose {
				fmt.Fprintf(os.Stderr, "  skip %s: %v\n", entry.Name(), err)
			}
			continue
		}
		if l.Cfg.Verbose {
			fmt.Printf("  imported %s -> %s [%s]\n", fullPath, target, category)
		}
		count++
	}
	return count, nil
}

func splitPath(p string) []string {
	var parts []string
	for p != "" && p != "." && p != "/" {
		dir, file := filepath.Split(p)
		if file != "" {
			parts = append([]string{file}, parts...)
		}
		p = filepath.Clean(dir)
	}
	return parts
}
