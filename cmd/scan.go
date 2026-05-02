// Copyright 2026 Tim Jespers
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/tjespers/dofs/internal/model"
	"github.com/tjespers/dofs/internal/scanner"

	"github.com/spf13/cobra"
)

var (
	scanAdopt    bool
	scanAuto     bool
	scanIgnore   string
	scanCategory string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan home directory for adoptable dotfiles",
	Long: `Scan ~/ for dotfiles and dotfolders that are not yet managed by DOFS.
Use --adopt for interactive adoption or --auto to adopt all.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openStore(); err != nil {
			return err
		}
		defer closeStore()

		s := &scanner.Scanner{Cfg: cfg, Store: store}

		// Handle --ignore
		if scanIgnore != "" {
			if err := s.Ignore(scanIgnore); err != nil {
				return fmt.Errorf("add ignore pattern: %w", err)
			}
			fmt.Printf("Added ignore pattern: %s\n", scanIgnore)
			return nil
		}

		candidates, err := s.FindCandidates()
		if err != nil {
			return err
		}

		if len(candidates) == 0 {
			fmt.Println("No adoptable dotfiles found.")
			return nil
		}

		category := model.CategoryHome
		if scanCategory != "" {
			category = model.Category(scanCategory)
		}

		// --adopt: interactive
		if scanAdopt {
			n, err := s.AdoptInteractive(candidates, category)
			if err != nil {
				return err
			}
			fmt.Printf("Adopted %d item(s)\n", n)
			return nil
		}

		// --auto: adopt all
		if scanAuto {
			adopted := 0
			for _, c := range candidates {
				if err := s.Adopt(c, category); err != nil {
					fmt.Fprintf(os.Stderr, "  error adopting %s: %v\n", c.Name, err)
					continue
				}
				fmt.Printf("  adopted %s\n", c.Name)
				adopted++
			}
			fmt.Printf("Adopted %d item(s)\n", adopted)
			return nil
		}

		// Default: list candidates
		fmt.Printf("Found %d adoptable dotfile(s):\n", len(candidates))
		for _, c := range candidates {
			kind := "file"
			if c.IsDir {
				kind = "dir "
			}
			fmt.Printf("  %s  ~/%s\n", kind, c.Name)
		}
		fmt.Println("\nRun 'dofs scan --adopt' to interactively adopt them.")
		return nil
	},
}

func init() {
	scanCmd.Flags().BoolVar(&scanAdopt, "adopt", false, "interactively prompt to adopt each candidate")
	scanCmd.Flags().BoolVar(&scanAuto, "auto", false, "adopt all candidates without prompting")
	scanCmd.Flags().StringVar(&scanIgnore, "ignore", "", "add a pattern to the persistent ignore list")
	scanCmd.Flags().StringVar(&scanCategory, "category", "", "target category for adopted items (default: home)")
	rootCmd.AddCommand(scanCmd)
}
