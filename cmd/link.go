// Copyright 2026 Tim Jespers
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/tjespers/dofs/internal/linker"
	"github.com/tjespers/dofs/internal/model"

	"github.com/spf13/cobra"
)

var (
	linkCreate   bool
	linkRepair   bool
	linkForce    bool
	linkCategory string
	linkAdd      string
	linkTarget   string
	linkRemove   int64
)

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "Manage symlinks between DOFS and home directory",
	Long: `Verify, create, repair, or manage tracked symlinks.

Without flags, shows the status of all tracked symlinks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := openStore(); err != nil {
			return err
		}
		defer closeStore()

		l := &linker.Linker{Cfg: cfg, Store: store}

		var cat *model.Category
		if linkCategory != "" {
			c := model.Category(linkCategory)
			cat = &c
		}

		// Handle --add
		if linkAdd != "" {
			category := model.CategoryHome
			if linkCategory != "" {
				category = model.Category(linkCategory)
			}
			if err := l.Add(category, linkAdd, linkTarget); err != nil {
				return fmt.Errorf("add link: %w", err)
			}
			fmt.Printf("Added %s (category: %s)\n", linkAdd, category)
			if linkCreate {
				n, err := l.CreateMissing(cat, linkForce)
				if err != nil {
					return err
				}
				fmt.Printf("Created %d symlink(s)\n", n)
			}
			return nil
		}

		// Handle --remove
		if linkRemove > 0 {
			if err := l.Remove(linkRemove); err != nil {
				return fmt.Errorf("remove link: %w", err)
			}
			fmt.Printf("Removed link #%d from tracking\n", linkRemove)
			return nil
		}

		// Handle --repair
		if linkRepair {
			n, err := l.RepairBroken(cat)
			if err != nil {
				return err
			}
			fmt.Printf("Repaired %d symlink(s)\n", n)
			return nil
		}

		// Handle --create
		if linkCreate {
			n, err := l.CreateMissing(cat, linkForce)
			if err != nil {
				return err
			}
			fmt.Printf("Created %d symlink(s)\n", n)
			return nil
		}

		// Default: show status
		results, err := l.CheckAll(cat)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Println("No tracked links. Run 'dofs init' to import existing symlinks.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "STATUS\tCATEGORY\tSOURCE\tTARGET\n")
		for _, r := range results {
			status := statusIcon(r.Status)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", status, r.Link.Category, r.Link.SourceName, r.Link.TargetPath)
		}
		w.Flush()
		return nil
	},
}

func statusIcon(s model.LinkStatus) string {
	switch s {
	case model.LinkOK:
		return "ok"
	case model.LinkMissing:
		return "MISSING"
	case model.LinkBroken:
		return "BROKEN"
	case model.LinkConflict:
		return "CONFLICT"
	case model.LinkWrongTarget:
		return "WRONG"
	default:
		return "?"
	}
}

func init() {
	linkCmd.Flags().BoolVar(&linkCreate, "create", false, "create missing symlinks")
	linkCmd.Flags().BoolVar(&linkRepair, "repair", false, "fix broken symlinks")
	linkCmd.Flags().BoolVar(&linkForce, "force", false, "overwrite conflicts (backs up existing files)")
	linkCmd.Flags().StringVar(&linkCategory, "category", "", "filter by category (home, secrets)")
	linkCmd.Flags().StringVar(&linkAdd, "add", "", "register a new item to track")
	linkCmd.Flags().StringVar(&linkTarget, "target", "", "custom target path (use with --add)")
	linkCmd.Flags().Int64Var(&linkRemove, "remove", 0, "stop tracking a link by ID")
	rootCmd.AddCommand(linkCmd)
}
