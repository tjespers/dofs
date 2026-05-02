// Copyright 2026 Tim Jespers
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Root    string
	DBPath  string
	HomeDir string
	Verbose bool
	DryRun  bool
}

func DefaultRoot() string {
	if v := os.Getenv("DOFS_ROOT"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "dofs", "data")
}

func Load(rootFlag, dbFlag string, verbose, dryRun bool) (*Config, error) {
	root := rootFlag
	if root == "" {
		root = DefaultRoot()
	}

	// Expand ~ prefix
	if len(root) > 0 && root[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot resolve home directory: %w", err)
		}
		root = filepath.Join(home, root[1:])
	}

	var err error
	root, err = filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve root path: %w", err)
	}

	dbPath := dbFlag
	if dbPath == "" {
		dbPath = filepath.Join(root, ".dofs.db")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot resolve home directory: %w", err)
	}

	return &Config{
		Root:    root,
		DBPath:  dbPath,
		HomeDir: home,
		Verbose: verbose,
		DryRun:  dryRun,
	}, nil
}
