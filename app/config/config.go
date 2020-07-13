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

// HistoryFile is the name of the file storing command line history
var HistoryFile = "history"

// MaxNameLen is the maximum length for entry identifier values
var MaxNameLen = 50

// Prompt is used in readline mode
var Prompt = "\033[1;32mmemory\033[0m> "

// SubPrompt is used within an interactive command loop
var SubPrompt = ": "

// TruncateAt is the length that values are truncated to with an ... during display
var TruncateAt = 300

// EditorCommand is the command to launch an external editor for long text values
//TODO: handle editor command cross-platform
var EditorCommand = "/usr/bin/vim"

// SavePath returns the full path to the data file
func SavePath() string {
	return MemoryHome + slash + DataFile
}

// HistoryPath returns the full path to the history file
func HistoryPath() string {
	return MemoryHome + slash + HistoryFile
}
