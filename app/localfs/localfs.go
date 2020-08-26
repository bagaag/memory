/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/* Contains functions that use the local file system. */

package localfs

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"memory/app/config"
	"memory/util"
	"os"
	"sync"
	"time"
)

var lock sync.Mutex

// marshal the object into an io.Reader
func marshal(v interface{}) (io.Reader, error) {
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
	r, err := marshal(v)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return err
}

// unmarshal data from the reader into the specified value
func unmarshal(r io.Reader, v interface{}) error {
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
	return unmarshal(f, v)
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
	if _, err = w.WriteString(content); err != nil {
		return tempFile.Name(), err
	}
	if err = w.Flush(); err != nil {
		return tempFile.Name(), err
	}

	return tempFile.Name(), err
}

// ReadFile returns the string contents of the text file.
func ReadFile(path string) (string, time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", time.Now(), err
	}
	bs, err := ioutil.ReadFile(path)
	return string(bs), info.ModTime(), err
}

// RemoveFile deletes the specified file.
func RemoveFile(path string) error {
	return os.Remove(path)
}

// InitHome checks that the home, entries and temp folders exist and creates them if needed.
func InitHome() error {
	if !util.PathExists(config.MemoryHome) {
		err := os.MkdirAll(config.EntriesPath(), 0740)
		if err != nil {
			fmt.Println("Failed to initialize settings folder at", config.MemoryHome)
			panic(err)
		}
	}
	if !util.PathExists(config.EntriesPath()) {
		err := os.MkdirAll(config.EntriesPath(), 0740)
		if err != nil {
			fmt.Println("Failed to initialize entries folder at", config.EntriesPath())
			panic(err)
		}
	}
	if !util.PathExists(config.TempPath()) {
		err := os.MkdirAll(config.TempPath(), 0740)
		if err != nil {
			fmt.Println("Failed to initialize temp folder at", config.TempPath())
			panic(err)
		}
	}
	if !util.PathExists(config.FilesPath()) {
		err := os.MkdirAll(config.FilesPath(), 0740)
		if err != nil {
			fmt.Println("Failed to initialize files folder at", config.FilesPath())
			panic(err)
		}
	}
	return nil
}
