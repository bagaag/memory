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

var MemoryHome = "~/.memory"
var DataFile = "memory.json"

func SavePath() string {
	return MemoryHome + slash + DataFile
}
