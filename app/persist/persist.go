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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"memory/app/config"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var lock sync.Mutex

// Marshal the object into an io.Reader
func Marshal(v interface{}) (io.Reader, error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

// Save saves a representation of v to the file at path
func Save(path string, v interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r, err := Marshal(v)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return err
}

// Unmarshal data from the reader into the specified value
func Unmarshal(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// Load the json file at path into v
func Load(path string, v interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return Unmarshal(f, v)
}

// PathExists returns true if the given path exists.
func PathExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		}
	}
	return true
}

// CreateTempFile returns the full path to a new temporary file containing
// a copy of the source file identified by the given slug.
func CreateTempFile(slug string, content string) (string, error) {
	var tempFile *os.File
	var err error
	// TODO: Clean up temp files older than 24 hrs at startup

	// temp file we'll write to and return the name of
	if tempFile, err = ioutil.TempFile(config.TempPath(), slug+"-*"+config.EntryExt); err != nil {
		return "", err
	}
	defer tempFile.Close()

	w := bufio.NewWriter(tempFile)
	w.WriteString(content)
	w.Flush()

	return tempFile.Name(), err
}

// slugToStoragePath converts a slug into a storage path.
func slugToStoragePath(slug string) string {
	return config.EntriesPath() + config.Slash + slug + config.EntryExt
}

// EntryExists returns true if the given slug is backed by physical storage.
func EntryExists(slug string) bool {
	return PathExists(slugToStoragePath(slug))
}

// ReadEntry converts a slug into a storage path and returns the source data for the entry.
func ReadEntry(slug string) (string, time.Time, error) {
	path := slugToStoragePath(slug)
	if !PathExists(path) {
		return "", time.Now(), fmt.Errorf("source file for %s not found", path)
	}
	return ReadFile(path)
}

// ReadFile returns the string contents of the text file.
func ReadFile(path string) (string, time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", time.Now(), err
	}
	bytes, err := ioutil.ReadFile(path)
	return string(bytes), info.ModTime(), err
}

// RemoveFile deletes the temporary editing file.
func RemoveFile(path string) error {
	return os.Remove(path)
}

// InitHome checks that the home, entries and temp folders exist and creates them if needed.
func InitHome() error {
	if !PathExists(config.MemoryHome) {
		err := os.MkdirAll(config.EntriesPath(), 0740)
		if err != nil {
			fmt.Println("Failed to initialize settings folder at", config.MemoryHome)
			panic(err)
		}
	}
	if !PathExists(config.EntriesPath()) {
		err := os.MkdirAll(config.EntriesPath(), 0740)
		if err != nil {
			fmt.Println("Failed to initialize entries folder at", config.EntriesPath())
			panic(err)
		}
	}
	if !PathExists(config.TempPath()) {
		err := os.MkdirAll(config.TempPath(), 0740)
		if err != nil {
			fmt.Println("Failed to initialize temp folder at", config.TempPath())
			panic(err)
		}
	}
	return nil
}

// EntrySlugs returns a string slice of entry file paths
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

// EntryFileName returns the storage identifier for an entry given the slug
func EntryFileName(slug string) string {
	return config.EntriesPath() + config.Slash + slug + config.EntryExt
}

// SaveEntry saves the text content of an entry to storage
func SaveEntry(slug string, content string) error {
	f, err := os.Create(EntryFileName(slug))
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	w.WriteString(content)
	w.Flush()
	return err
}

// DeleteEntry deletes the entry identified by the slug
func DeleteEntry(slug string) error {
	return os.Remove(EntryFileName(slug))
}
