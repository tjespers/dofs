// Copyright 2026 Tim Jespers
// SPDX-License-Identifier: Apache-2.0

package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/tjespers/dofs/internal/config"
	"github.com/tjespers/dofs/internal/db"
	"github.com/tjespers/dofs/internal/model"
)

type Archiver struct {
	Cfg            *config.Config
	Store          *db.Store
	IncludeAll     bool // if true, skip no directories
}

func (a *Archiver) BackupCategory(cat model.Category, outputDir string) (*model.BackupRecord, error) {
	srcDir := filepath.Join(a.Cfg.Root, string(cat))
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("category directory does not exist: %s", srcDir)
	}

	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("dofs-backup-%s-%s.tar.gz", cat, timestamp)
	outPath := filepath.Join(outputDir, filename)

	rec := &model.BackupRecord{Category: cat, ArchivePath: outPath}
	recID, err := a.Store.InsertBackup(rec)
	if err != nil {
		return nil, fmt.Errorf("record backup: %w", err)
	}

	if a.Cfg.DryRun {
		fmt.Printf("[dry-run] would create %s from %s\n", outPath, srcDir)
		_ = a.Store.FailBackup(recID)
		rec.Status = "dry-run"
		return rec, nil
	}

	f, err := os.Create(outPath)
	if err != nil {
		_ = a.Store.FailBackup(recID)
		return nil, fmt.Errorf("create archive: %w", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	err = filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && !a.IncludeAll && DefaultExcludes[d.Name()] {
			if a.Cfg.Verbose {
				fmt.Printf("  excluding %s\n", path)
			}
			return fs.SkipDir
		}

		// Get relative path for the archive (contents only, no root dir name)
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		// Skip sockets, devices, and other special files
		if info.Mode()&os.ModeSocket != 0 || info.Mode()&os.ModeDevice != 0 ||
			info.Mode()&os.ModeNamedPipe != 0 {
			return nil
		}

		// Handle symlinks
		var link string
		if info.Mode()&os.ModeSymlink != 0 {
			link, err = os.Readlink(path)
			if err != nil {
				return err
			}
		}

		header, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}
		header.Name = rel

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !d.IsDir() && info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		_ = a.Store.FailBackup(recID)
		os.Remove(outPath)
		return nil, fmt.Errorf("archive %s: %w", cat, err)
	}

	// Close writers before stat to flush
	tw.Close()
	gw.Close()

	fi, err := os.Stat(outPath)
	if err != nil {
		_ = a.Store.FailBackup(recID)
		return nil, err
	}

	_ = a.Store.CompleteBackup(recID, fi.Size())
	rec.SizeBytes = fi.Size()
	rec.Status = "completed"
	return rec, nil
}

func (a *Archiver) BackupAll(categories []model.Category, outputDir string) ([]*model.BackupRecord, error) {
	var records []*model.BackupRecord
	for _, cat := range categories {
		rec, err := a.BackupCategory(cat, outputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  error backing up %s: %v\n", cat, err)
			continue
		}
		records = append(records, rec)
	}
	return records, nil
}

func ListExcludes() []string {
	excludes := make([]string, 0, len(DefaultExcludes))
	for k := range DefaultExcludes {
		excludes = append(excludes, k)
	}
	return excludes
}
