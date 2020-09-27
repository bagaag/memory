/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package persist

import (
	"bytes"
	"encoding/json"
	"io"
	"memory/app/config"
	"memory/app/localfs"
	"memory/app/model"
	"memory/util"
	"os"
	"path/filepath"
	"strings"
)

// Config struct for SimplePersist
//TODO: Implement https://github.com/uber-go/config
type SimplePersistConfig struct {
	EntryPath string
	FilePath  string
}

// Implementation of the Persist interface that uses the local file system.
type SimplePersist struct {
	cfg   SimplePersistConfig
	slash string
	ext   string
}

// Creates and configures a new instance of SimplePersist.
func NewSimplePersist(cfg SimplePersistConfig) (SimplePersist, error) {
	p := SimplePersist{cfg: cfg, slash: string(os.PathSeparator), ext: ".json"}
	if !localfs.PathExists(p.cfg.EntryPath) {
		err := os.MkdirAll(p.cfg.EntryPath, 0740)
		if err != nil {
			return p, err
		}
	}
	if !localfs.PathExists(p.cfg.FilePath) {
		err := os.MkdirAll(p.cfg.FilePath, 0740)
		if err != nil {
			return p, err
		}
	}
	return p, nil
}

// EntryExists returns true if the given slug matches an indexed entry.
func (p *SimplePersist) EntryExists(slug string) bool {
	path := p.slugToStoragePath(slug)
	return localfs.PathExists(path)
}

// ReadEntry returns an Entry identified by slug populated from storage.
func (p *SimplePersist) ReadEntry(slug string) (model.Entry, error) {
	path := p.slugToStoragePath(slug)
	if !localfs.PathExists(path) {
		return model.Entry{}, model.EntryNotFound{Slug: slug}
	}
	var entry model.Entry
	err := p.load(path, &entry)
	if err != nil {
		return entry, err
	}
	entry.SetPopulated(true)
	return entry, nil
}

// EntrySlugs returns a string slice containing the slug of every entry in storage.
func (p *SimplePersist) EntrySlugs() ([]string, error) {
	paths, err := filepath.Glob(p.cfg.EntryPath + config.Slash + "*" + p.ext)
	if err != nil {
		return []string{}, err
	}
	for ix, path := range paths {
		parts := strings.Split(path, p.slash)
		path = parts[len(parts)-1]
		path = strings.TrimSuffix(path, p.ext)
		paths[ix] = path
	}
	return paths, nil
}

// SaveEntry writes the entry to storage.
func (p *SimplePersist) SaveEntry(entry model.Entry) error {
	path := p.slugToStoragePath(entry.Slug())
	return p.save(path, entry)
}

// DeleteEntry removes the entry idenfied by slug from storage.
func (p *SimplePersist) DeleteEntry(slug string) error {
	path := p.slugToStoragePath(slug)
	return os.Remove(path)
}

// RenameEntry moves an entry from one slug to another, reflecting a new name and
// returning the slug for the renamed entry
func (p *SimplePersist) RenameEntry(oldName string, newName string) (model.Entry, error) {
	oldSlug := util.GetSlug(oldName)
	entry, err := p.ReadEntry(oldSlug)
	if err != nil {
		return model.Entry{}, err
	}
	entry.Name = newName
	if err = p.SaveEntry(entry); err != nil {
		return model.Entry{}, err
	}
	if err = p.DeleteEntry(oldSlug); err != nil {
		return entry, err
	}
	return entry, nil
}

// slugToStoragePath converts a slug into a storage path.
func (p *SimplePersist) slugToStoragePath(slug string) string {
	return p.cfg.EntryPath + p.slash + slug + p.ext
}

// Marshal the object into an io.Reader
func (p *SimplePersist) marshal(v interface{}) (io.Reader, error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

// Save saves a representation of v to the file at path
func (p *SimplePersist) save(path string, v interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r, err := p.marshal(v)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return err
}

// Unmarshal data from the reader into the specified value
func (p *SimplePersist) unmarshal(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// Load the json file at path into v
func (p *SimplePersist) load(path string, v interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return p.unmarshal(f, v)
}
