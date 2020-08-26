/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

// The persist package handles persistence tasks.

package persist

import (
	"bufio"
	"fmt"
	config "memory/app/config"
	"memory/app/localfs"
	"memory/app/model"
	"memory/app/template"
	util "memory/util"
	"os"
	"path/filepath"
	"strings"
)

type EntryNotFound struct {
	Slug string
}

func (e EntryNotFound) Error() string {
	return fmt.Sprintf("entry %s not found", e.Slug)
}

// slugToStoragePath converts a slug into a storage path.
func slugToStoragePath(slug string) string {
	return config.EntriesPath() + config.Slash + slug + config.EntryExt
}

// ReadEntry converts a slug into a storage path and returns the source data for the entry.
func ReadEntry(slug string) (model.Entry, error) {
	path := slugToStoragePath(slug)
	if !util.PathExists(path) {
		return model.Entry{}, EntryNotFound{slug}
	}
	content, modified, err := localfs.ReadFile(path)
	entry, err := template.ParseYamlDown(content)
	if err != nil {
		return model.Entry{}, err
	}
	entry.Modified = modified
	return entry, nil
}

// EntrySlugs returns a string slice of entry slugs found in storage.
func EntrySlugs() ([]string, error) {
	paths, err := filepath.Glob(config.EntriesPath() + config.Slash + "*" + config.EntryExt)
	if err != nil {
		return []string{}, err
	}
	for ix, path := range paths {
		parts := strings.Split(path, config.Slash)
		path = parts[len(parts)-1]
		path = strings.TrimSuffix(path, config.EntryExt)
		paths[ix] = path
	}
	return paths, nil
}

// entryFileName returns the storage identifier for an entry given the slug
func entryFileName(slug string) string {
	return config.EntriesPath() + config.Slash + slug + config.EntryExt
}

// SaveEntry saves the text content of an entry to storage
func SaveEntry(slug string, content string) error {
	f, err := os.Create(entryFileName(slug))
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	if _, err = w.WriteString(content); err != nil {
		return err
	}
	if err = w.Flush(); err != nil {
		return err
	}
	return nil
}

// DeleteEntry deletes the entry identified by the slug
func DeleteEntry(slug string) error {
	return os.Remove(entryFileName(slug))
}
