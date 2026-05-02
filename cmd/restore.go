// Copyright 2026 Tim Jespers
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjespers/dofs/internal/linker"
	"github.com/tjespers/dofs/internal/model"

	"github.com/spf13/cobra"
)

var (
	restoreCreateLinks bool
)

var restoreCmd = &cobra.Command{
	Use:   "restore <archive.tar.gz>...",
	Short: "Restore DOFS from backup archives",
	Long: `Restore one or more backup archives into the DOFS root, initialize the
database, register all items in linkable categories, and optionally create symlinks.

Archive filenames are expected to follow the naming convention:
  dofs-backup-<category>-<timestamp>.tar.gz

The category is extracted from the filename and used as the target directory.

Example:
  dofs restore dofs-backup-home-*.tar.gz dofs-backup-secrets-*.tar.gz dofs-backup-code-*.tar.gz
  dofs restore --link *.tar.gz`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := loadConfig(); err != nil {
			return err
		}

		// Ensure DOFS root exists
		if err := os.MkdirAll(cfg.Root, 0755); err != nil {
			return fmt.Errorf("create DOFS root: %w", err)
		}

		// Extract each archive
		for _, archive := range args {
			category, err := categoryFromFilename(archive)
			if err != nil {
				return err
			}

			destDir := filepath.Join(cfg.Root, category)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return fmt.Errorf("create %s: %w", destDir, err)
			}

			if cfg.DryRun {
				fmt.Printf("[dry-run] would extract %s -> %s\n", archive, destDir)
				continue
			}

			fmt.Printf("Extracting %s -> %s\n", filepath.Base(archive), destDir)
			n, err := extractArchive(archive, destDir)
			if err != nil {
				return fmt.Errorf("extract %s: %w", archive, err)
			}
			fmt.Printf("  %d items extracted\n", n)
		}

		if cfg.DryRun {
			return nil
		}

		// Initialize DB
		if err := openStoreWithConfig(); err != nil {
			return fmt.Errorf("initialize database: %w", err)
		}
		defer closeStore()

		fmt.Printf("\nDatabase initialized at %s\n", cfg.DBPath)

		// Register all items in linkable categories
		l := &linker.Linker{Cfg: cfg, Store: store}
		registered := 0
		for _, cat := range model.LinkableCategories {
			dir := filepath.Join(cfg.Root, string(cat))
			entries, err := os.ReadDir(dir)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return fmt.Errorf("read %s: %w", dir, err)
			}
			for _, e := range entries {
				targetPath := filepath.Join(cfg.HomeDir, e.Name())

				if err := l.Add(cat, e.Name(), targetPath); err != nil {
					if cfg.Verbose {
						fmt.Fprintf(os.Stderr, "  skip %s/%s: %v\n", cat, e.Name(), err)
					}
					continue
				}
				if cfg.Verbose {
					fmt.Printf("  registered %s/%s -> %s\n", cat, e.Name(), targetPath)
				}
				registered++
			}
		}
		fmt.Printf("Registered %d items from linkable categories\n", registered)

		// Optionally create symlinks
		if restoreCreateLinks {
			n, err := l.CreateMissing(nil, false)
			if err != nil {
				return fmt.Errorf("create symlinks: %w", err)
			}
			fmt.Printf("Created %d symlink(s)\n", n)
		} else if registered > 0 {
			fmt.Println("\nRun 'dofs link --create' to create symlinks, or use 'dofs link' to review first.")
		}

		return nil
	},
}

// categoryFromFilename extracts the category from a dofs backup filename.
// Expected format: dofs-backup-<category>-<timestamp>.tar.gz
// Also accepts: <category>.tar.gz as a fallback
func categoryFromFilename(path string) (string, error) {
	base := filepath.Base(path)

	// Try standard naming: dofs-backup-<category>-YYYYMMDD-HHMMSS.tar.gz
	if strings.HasPrefix(base, "dofs-backup-") {
		rest := strings.TrimPrefix(base, "dofs-backup-")
		// Find the category part (everything before the timestamp)
		// Timestamp format: YYYYMMDD-HHMMSS
		parts := strings.Split(rest, "-")
		if len(parts) >= 3 {
			// Category could be multi-word, timestamp is last 2 parts (date-time) before .tar.gz
			// But our categories are single words, so first part is the category
			category := parts[0]
			return category, nil
		}
	}

	return "", fmt.Errorf("cannot determine category from filename %q — expected dofs-backup-<category>-<timestamp>.tar.gz", base)
}

func extractArchive(archivePath, destDir string) (int, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return 0, err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	count := 0

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}

		target := filepath.Join(destDir, header.Name)

		// Protect against zip-slip
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(destDir)+string(os.PathSeparator)) &&
			filepath.Clean(target) != filepath.Clean(destDir) {
			return count, fmt.Errorf("archive entry %q attempts path traversal", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return count, err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return count, err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return count, err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return count, err
			}
			outFile.Close()
			count++
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return count, err
			}
			// Remove existing symlink if any
			os.Remove(target)
			if err := os.Symlink(header.Linkname, target); err != nil {
				return count, err
			}
			count++
		}
	}

	return count, nil
}

func init() {
	restoreCmd.Flags().BoolVar(&restoreCreateLinks, "link", false, "create symlinks after restoring")
	rootCmd.AddCommand(restoreCmd)
}
