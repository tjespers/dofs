// Copyright 2026 Tim Jespers
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/tjespers/dofs/internal/linker"
	"github.com/tjespers/dofs/internal/model"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize DOFS database and import existing symlinks",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := loadConfig(); err != nil {
			return err
		}

		// Verify root exists
		info, err := os.Stat(cfg.Root)
		if err != nil || !info.IsDir() {
			return fmt.Errorf("DOFS root does not exist or is not a directory: %s", cfg.Root)
		}

		// Check for expected subdirectories
		fmt.Printf("DOFS root: %s\n", cfg.Root)
		for _, cat := range model.AllCategories {
			dir := cfg.Root + "/" + string(cat)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				fmt.Printf("  warning: %s/ does not exist\n", cat)
			}
		}

		// Open DB (creates and migrates if needed)
		if err := openStoreWithConfig(); err != nil {
			return fmt.Errorf("initialize database: %w", err)
		}
		defer closeStore()

		fmt.Printf("Database: %s\n", cfg.DBPath)

		// Import existing symlinks
		l := &linker.Linker{Cfg: cfg, Store: store}
		count, err := l.ImportExisting()
		if err != nil {
			return fmt.Errorf("import existing symlinks: %w", err)
		}
		fmt.Printf("Imported %d existing symlinks\n", count)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
