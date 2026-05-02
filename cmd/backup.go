// Copyright 2026 Tim Jespers
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"sort"

	"github.com/tjespers/dofs/internal/backup"
	"github.com/tjespers/dofs/internal/model"

	"github.com/spf13/cobra"
)

var (
	backupOutput      string
	backupListExcl    bool
	backupIncludeAll  bool
)

var backupCmd = &cobra.Command{
	Use:   "backup [category...]",
	Short: "Create compressed archives of DOFS categories",
	Long: `Archive DOFS categories as .tar.gz files, excluding package manager
and build output directories by default.

Without arguments, backs up all categories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if backupListExcl {
			excludes := backup.ListExcludes()
			sort.Strings(excludes)
			fmt.Println("Excluded directories:")
			for _, e := range excludes {
				fmt.Printf("  %s\n", e)
			}
			return nil
		}

		if err := openStore(); err != nil {
			return err
		}
		defer closeStore()

		outputDir := backupOutput
		if outputDir == "" {
			outputDir = "."
		}

		a := &backup.Archiver{
			Cfg:        cfg,
			Store:      store,
			IncludeAll: backupIncludeAll,
		}

		var categories []model.Category
		if len(args) > 0 {
			for _, arg := range args {
				categories = append(categories, model.Category(arg))
			}
		} else {
			categories = model.AllCategories
		}

		records, err := a.BackupAll(categories, outputDir)
		if err != nil {
			return err
		}

		for _, rec := range records {
			if rec.Status == "dry-run" {
				continue
			}
			fmt.Printf("  %s: %s (%.1f MB)\n", rec.Category, rec.ArchivePath,
				float64(rec.SizeBytes)/1024/1024)
		}
		return nil
	},
}

func init() {
	backupCmd.Flags().StringVarP(&backupOutput, "output", "o", "", "output directory (default: current directory)")
	backupCmd.Flags().BoolVar(&backupListExcl, "list-excludes", false, "show exclusion patterns and exit")
	backupCmd.Flags().BoolVar(&backupIncludeAll, "include-all", false, "do not exclude package manager directories")
	rootCmd.AddCommand(backupCmd)
}
