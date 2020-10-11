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

// StoredSettings are the settings written to the settings.json file in MemoryHome/.
type StoredSettings struct {
	EditorCommand string
}

const Version = "1.0"

const Slash = string(os.PathSeparator)

// MemoryHome is the folder path where memory stores settings and data
var MemoryHome = ".memory"

// EntryDir is the folder path where entry files are stored
var EntryDir = "entries"

// DataFile is the name of the file storing entries
var DataFile = "memory.json"

// SettingsFile is the name of the file for storing preferences
var SettingsFile = "settings.json"

// HistoryFile is the name of the file storing command line history
var HistoryFile = "history.txt"

// OpenFileCommand is the command to use when opening an attached file
//TODO: handle cross-platform commmand
//Linux: xdg-open"
//Win: start "" "%"
//Mac: open "%"
var OpenFileCommand = "xdg-open"

// SettingsFile is the name of the file storing the settings struct

// MaxNameLen is the maximum length for entry identifier values
var MaxNameLen = 50

// Prompt is used in readline mode
var Prompt = "\033[1;32mmemory\033[0m> "

// SubPrompt is used within an interactive command loop
var SubPrompt = ": "

// EditorCommand is the command to launch an external editor for long text values
//TODO: handle editor command cross-platform
var EditorCommand = "/usr/bin/micro"

// EntryExt is the file extension (including .) used for entry files
var EntryExt = ".txt"

// SavePath returns the full path to the data file
func SavePath() string {
	return MemoryHome + Slash + DataFile
}

// HistoryPath returns the full path to the history file
func HistoryPath() string {
	return MemoryHome + Slash + HistoryFile
}

// SettingsPath returns the full path to the settings file
func SettingsPath() string {
	return MemoryHome + Slash + SettingsFile
}

// EntriesPath returns the full path to EntryDir
func EntriesPath() string {
	return MemoryHome + Slash + EntryDir
}

// TempPath returns the location where temporary files are stored during editing.
func TempPath() string {
	return MemoryHome + Slash + "tmp"
}

// GetSettingsForStorage returns a StoredSettings struct populated with current settings.
func GetSettingsForStorage() StoredSettings {
	settings := StoredSettings{
		EditorCommand: EditorCommand,
	}
	return settings
}

// UpdateSettingsFromStorage sets active settings from a populated StoredSettings object.
func UpdateSettingsFromStorage(settings StoredSettings) {
	EditorCommand = settings.EditorCommand
}

// SearchPath returns the full path to the search index database
func SearchPath() string {
	return MemoryHome + Slash + "search.bleve"
}

// FilesPath returns the full path to the files folder where attachments are stored.
func FilesPath() string {
	return MemoryHome + Slash + "files"
}
