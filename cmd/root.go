// Copyright 2026 Tim Jespers
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/tjespers/dofs/internal/config"
	"github.com/tjespers/dofs/internal/db"

	"github.com/spf13/cobra"
)

var (
	cfgRootFlag    string
	cfgDBFlag      string
	cfgVerbose     bool
	cfgDryRun      bool
	cfg            *config.Config
	store          *db.Store
)

var rootCmd = &cobra.Command{
	Use:   "dofs",
	Short: "Developer Optimized File System — manage your workstation files",
	Long: `DOFS manages a structured directory tree with symlinks, backups, and
dotfile adoption. Keep your workstation portable and reproducible.`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgRootFlag, "root", "", fmt.Sprintf("DOFS root directory (default: $DOFS_ROOT or %s)", config.DefaultRoot()))
	rootCmd.PersistentFlags().StringVar(&cfgDBFlag, "db", "", "SQLite database path (default: <root>/.dofs.db)")
	rootCmd.PersistentFlags().BoolVarP(&cfgVerbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&cfgDryRun, "dry-run", false, "show what would be done without doing it")
}

func loadConfig() error {
	var err error
	cfg, err = config.Load(cfgRootFlag, cfgDBFlag, cfgVerbose, cfgDryRun)
	return err
}

func openStore() error {
	if err := loadConfig(); err != nil {
		return err
	}
	var err error
	store, err = db.Open(cfg.DBPath)
	return err
}

// openStoreWithConfig opens the DB using an already-loaded config.
func openStoreWithConfig() error {
	var err error
	store, err = db.Open(cfg.DBPath)
	return err
}

func closeStore() {
	if store != nil {
		store.Close()
	}
}
