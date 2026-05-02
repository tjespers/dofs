// Copyright 2026 Tim Jespers
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/tjespers/dofs/internal/model"
	"github.com/tjespers/dofs/internal/status"

	"github.com/spf13/cobra"
)

var (
	statusJSON  bool
	statusQuiet bool
	statusLinks bool
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show DOFS health report",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openStore(); err != nil {
			return err
		}
		defer closeStore()

		r := &status.Reporter{Cfg: cfg, Store: store}
		report, err := r.Report()
		if err != nil {
			return err
		}

		if statusJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(report)
		}

		if statusQuiet {
			if report.Summary.BrokenCount+report.Summary.MissingCount+report.Summary.ConflictCount+report.Summary.WrongCount > 0 {
				os.Exit(1)
			}
			return nil
		}

		fmt.Printf("DOFS Status Report\n")
		fmt.Printf("==================\n")
		fmt.Printf("Root: %s\n\n", report.Root)

		// Symlinks summary
		fmt.Printf("Symlinks: %d tracked", report.Summary.TotalLinks)
		if report.Summary.TotalLinks > 0 {
			fmt.Printf(" — %d ok", report.Summary.OKCount)
			if report.Summary.BrokenCount > 0 {
				fmt.Printf(", %d broken", report.Summary.BrokenCount)
			}
			if report.Summary.MissingCount > 0 {
				fmt.Printf(", %d missing", report.Summary.MissingCount)
			}
			if report.Summary.ConflictCount > 0 {
				fmt.Printf(", %d conflict", report.Summary.ConflictCount)
			}
			if report.Summary.WrongCount > 0 {
				fmt.Printf(", %d wrong target", report.Summary.WrongCount)
			}
		}
		fmt.Println()

		if !statusLinks {
			// Broken links detail
			for _, lr := range report.Links {
				if lr.Status == model.LinkBroken {
					fmt.Printf("  BROKEN: %s -> %s\n", lr.Link.TargetPath, lr.Link.SourcePath)
				}
			}
			for _, lr := range report.Links {
				if lr.Status == model.LinkMissing {
					fmt.Printf("  MISSING: %s (symlink not created)\n", lr.Link.TargetPath)
				}
			}
			for _, lr := range report.Links {
				if lr.Status == model.LinkConflict {
					fmt.Printf("  CONFLICT: %s (regular file exists)\n", lr.Link.TargetPath)
				}
			}

			// Untracked items
			if len(report.UntrackedItems) > 0 {
				fmt.Printf("\nUntracked items in DOFS tree:\n")
				for _, item := range report.UntrackedItems {
					fmt.Printf("  %s\n", item)
				}
			}

			// Untracked symlinks
			if len(report.UntrackedLinks) > 0 {
				fmt.Printf("\nUntracked symlinks in ~/ pointing to DOFS:\n")
				for _, name := range report.UntrackedLinks {
					fmt.Printf("  ~/%s\n", name)
				}
			}

			// Backups
			fmt.Printf("\nBackups:\n")
			for _, cat := range model.AllCategories {
				b, ok := report.LastBackups[cat]
				if !ok || b == nil {
					fmt.Printf("  %-12s never\n", cat)
				} else if b.CompletedAt != nil {
					days := int(math.Round(time.Since(*b.CompletedAt).Hours() / 24))
					fmt.Printf("  %-12s %s (%d days ago)\n", cat, b.CompletedAt.Format("2006-01-02"), days)
				}
			}
		}

		return nil
	},
}

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "output as JSON")
	statusCmd.Flags().BoolVar(&statusQuiet, "quiet", false, "exit code only (0=healthy, 1=issues)")
	statusCmd.Flags().BoolVar(&statusLinks, "links", false, "only check symlinks")
	rootCmd.AddCommand(statusCmd)
}
