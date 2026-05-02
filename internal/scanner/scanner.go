// Copyright 2026 Tim Jespers
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjespers/dofs/internal/config"
	"github.com/tjespers/dofs/internal/db"
	"github.com/tjespers/dofs/internal/model"
)

// Directories/files that should never be suggested for adoption.
var builtinIgnores = map[string]bool{
	".cache": true, ".local": true, ".dbus": true, ".pki": true,
	".bash_history": true, ".zsh_history": true, ".lesshst": true,
	".wget-hsts": true, ".viminfo": true, ".sudo_as_admin_successful": true,
	".Xauthority": true, ".xsession-errors": true, ".ICEauthority": true,
	".motd_shown": true, ".landscape": true,
}

type Scanner struct {
	Cfg   *config.Config
	Store *db.Store
}

func (s *Scanner) FindCandidates() ([]model.ScanCandidate, error) {
	entries, err := os.ReadDir(s.Cfg.HomeDir)
	if err != nil {
		return nil, fmt.Errorf("read home dir: %w", err)
	}

	// Load tracked target paths
	links, err := s.Store.ListLinks(nil)
	if err != nil {
		return nil, err
	}
	tracked := make(map[string]bool)
	for _, l := range links {
		tracked[l.TargetPath] = true
	}

	// Load ignore patterns
	ignorePatterns, err := s.Store.ListIgnorePatterns()
	if err != nil {
		return nil, err
	}

	var candidates []model.ScanCandidate
	for _, e := range entries {
		name := e.Name()

		// Only dotfiles/dotfolders
		if !strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := filepath.Join(s.Cfg.HomeDir, name)

		// Skip if already tracked
		if tracked[fullPath] {
			continue
		}

		// Skip builtin ignores
		if builtinIgnores[name] {
			continue
		}

		// Skip user ignore patterns
		if matchesAny(name, ignorePatterns) {
			continue
		}

		// Skip if it's a symlink pointing into the DOFS root (already managed, just not tracked)
		info, err := os.Lstat(fullPath)
		if err != nil {
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(fullPath)
			if err == nil {
				if !filepath.IsAbs(target) {
					target = filepath.Join(filepath.Dir(fullPath), target)
				}
				target, _ = filepath.Abs(target)
				if strings.HasPrefix(target, s.Cfg.Root) {
					continue
				}
			}
		}

		var size int64
		if info.Mode().IsRegular() {
			size = info.Size()
		}

		candidates = append(candidates, model.ScanCandidate{
			Name:  name,
			Path:  fullPath,
			IsDir: info.IsDir() || info.Mode()&os.ModeSymlink != 0 && isSymlinkToDir(fullPath),
			Size:  size,
		})
	}

	return candidates, nil
}

func (s *Scanner) Adopt(candidate model.ScanCandidate, category model.Category) error {
	destDir := filepath.Join(s.Cfg.Root, string(category))
	destPath := filepath.Join(destDir, candidate.Name)

	// Ensure category directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create category dir: %w", err)
	}

	// Check destination doesn't already exist
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("destination already exists: %s", destPath)
	}

	if s.Cfg.DryRun {
		fmt.Printf("[dry-run] would move %s -> %s and create symlink\n", candidate.Path, destPath)
		return nil
	}

	// Move the file/dir into DOFS
	if err := os.Rename(candidate.Path, destPath); err != nil {
		return fmt.Errorf("move %s: %w", candidate.Name, err)
	}

	// Create symlink back
	if err := os.Symlink(destPath, candidate.Path); err != nil {
		// Try to move back on failure
		os.Rename(destPath, candidate.Path)
		return fmt.Errorf("create symlink: %w", err)
	}

	// Record in DB
	link := &model.ManagedLink{
		Category:   category,
		SourceName: candidate.Name,
		SourcePath: destPath,
		TargetPath: candidate.Path,
		LastStatus: model.LinkOK,
	}
	if err := s.Store.InsertLink(link); err != nil {
		return fmt.Errorf("record link: %w", err)
	}

	_ = s.Store.InsertScanResult(candidate.Path, model.ScanAdopted, &category)
	return nil
}

func (s *Scanner) Ignore(pattern string) error {
	return s.Store.AddIgnorePattern(pattern)
}

func (s *Scanner) AdoptInteractive(candidates []model.ScanCandidate, category model.Category) (int, error) {
	reader := bufio.NewReader(os.Stdin)
	adopted := 0

	for i, c := range candidates {
		kind := "file"
		if c.IsDir {
			kind = "dir"
		}
		fmt.Printf("[%d/%d] ~/%s (%s) — Adopt? [y/n/i(gnore)/q(uit)] ", i+1, len(candidates), c.Name, kind)

		input, err := reader.ReadString('\n')
		if err != nil {
			return adopted, err
		}
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "y", "yes":
			if err := s.Adopt(c, category); err != nil {
				fmt.Fprintf(os.Stderr, "  error: %v\n", err)
				continue
			}
			fmt.Printf("  adopted %s -> %s/%s\n", c.Name, category, c.Name)
			adopted++
		case "i", "ignore":
			if err := s.Ignore(c.Name); err != nil {
				fmt.Fprintf(os.Stderr, "  error: %v\n", err)
			}
			_ = s.Store.InsertScanResult(c.Path, model.ScanIgnored, nil)
			fmt.Printf("  ignored %s\n", c.Name)
		case "q", "quit":
			return adopted, nil
		default:
			// skip
		}
	}
	return adopted, nil
}

func matchesAny(name string, patterns []string) bool {
	for _, p := range patterns {
		if matched, _ := filepath.Match(p, name); matched {
			return true
		}
		if p == name {
			return true
		}
	}
	return false
}

func isSymlinkToDir(path string) bool {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return false
	}
	return info.IsDir()
}
