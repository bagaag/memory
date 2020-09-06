/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package impl

import (
	"bufio"
	"memory/app/config"
	"memory/app/localfs"
	"memory/app/model"
	"memory/app/template"
	"memory/util"
	"os"
	"path/filepath"
	"strings"
)

// Config struct for SimplePersist.
//TODO: Implement https://github.com/uber-go/config
type SimplePersistConfig struct {
	EntryPath string
	FilePath  string
	EntryExt  string
}

// Implementation of the Persist interface that uses the local file system.
type SimplePersist struct {
	cfg   SimplePersistConfig
	slash string
}

// Creates and configures a new instance of SimplePersist.
func NewSimplePersist(cfg SimplePersistConfig) (SimplePersist, error) {
	p := SimplePersist{cfg: cfg, slash: string(os.PathSeparator)}
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

// slugToStoragePath converts a slug into a storage path.
func (p *SimplePersist) slugToStoragePath(slug string) string {
	return p.cfg.EntryPath + p.slash + slug + p.cfg.EntryExt
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
		return model.Entry{}, model.EntryNotFound{slug}
	}
	content, modified, err := localfs.ReadFile(path)
	entry, err := template.ParseYamlDown(content)
	if err != nil {
		return model.Entry{}, err
	}
	entry.Modified = modified
	return entry, nil
}

// EntrySlugs returns a string slice containing the slug of every entry in storage.
func (p *SimplePersist) EntrySlugs() ([]string, error) {
	paths, err := filepath.Glob(p.cfg.EntryPath + config.Slash + "*" + p.cfg.EntryExt)
	if err != nil {
		return []string{}, err
	}
	for ix, path := range paths {
		parts := strings.Split(path, p.slash)
		path = parts[len(parts)-1]
		path = strings.TrimSuffix(path, p.cfg.EntryExt)
		paths[ix] = path
	}
	return paths, nil
}

// SaveEntry writes the entry to storage.
func (p *SimplePersist) SaveEntry(entry model.Entry) error {
	f, err := os.Create(entryFileName(entry.Slug()))
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	if content, err := template.RenderYamlDown(entry); err != nil {
		return err
	} else if _, err = w.WriteString(content); err != nil {
		return err
	} else if err = w.Flush(); err != nil {
		return err
	}
	return nil
}

// DeleteEntry removes the entry idenfied by slug from storage.
func (p *SimplePersist) DeleteEntry(slug string) error {
	return os.Remove(entryFileName(slug))
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

// entryFileName returns the storage identifier for an entry given the slug
func entryFileName(slug string) string {
	return config.EntriesPath() + config.Slash + slug + config.EntryExt
}
