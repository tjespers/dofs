// Copyright 2026 Tim Jespers
// SPDX-License-Identifier: Apache-2.0

package backup

var DefaultExcludes = map[string]bool{
	// JavaScript / Node
	"node_modules":  true,
	".next":         true,
	".nuxt":         true,
	".parcel-cache": true,
	"bower_components": true,

	// PHP
	"vendor": true,

	// Python
	".venv":         true,
	"venv":          true,
	"__pycache__":   true,
	".tox":          true,
	".eggs":         true,
	".mypy_cache":   true,
	".pytest_cache": true,

	// Java / Kotlin / Scala
	".gradle": true,
	".maven":  true,

	// Rust
	"target": true,

	// General build output
	"dist":   true,
	"build":  true,
	".cache": true,
}
