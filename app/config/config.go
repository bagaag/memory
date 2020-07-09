/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

// Package config stores app-specific settings. These should be set by
// the UI controller from a settings file, arguments or environment
// variables.
package config

import (
	"os"
)

const slash = string(os.PathSeparator)

// MemoryHome is the folder path where memory stores settings and data
var MemoryHome = "~/.memory"

// DataFile is the name of the file storing entries
var DataFile = "memory.json"

// MaxNameLen is the maximum length for entry identifier values
var MaxNameLen = 50

// SavePath returns the full path to the data file
func SavePath() string {
	return MemoryHome + slash + DataFile
}
