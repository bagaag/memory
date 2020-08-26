/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

//Package app contains an API for interacting with the application
//that is not bound to a particular UI.
package app

import (
	"fmt"
	"memory/app/config"
	"memory/app/localfs"
	"memory/app/model"
	"memory/app/persist"
	"memory/util"
	"os"
	"sort"
	"sync"
)

var inited = false // run Init() only once

var deletes = []string{}

// root contains all the data to be saved
type root struct {
	Names map[string]model.Entry
	mux   sync.Mutex
}

// EntryResults is used to contain the results of GetEntries and the settings used
// to generate those results.
type EntryResults struct {
	Entries  []model.Entry
	Types    model.EntryTypes
	Search   string
	AnyTags  []string
	OnlyTags []string
	Sort     SortOrder
	Total    uint64
	PageNo   int
	PageSize int
}

// SortOrder is used to indicate one of the Sort constants
type SortOrder int

// SortScore sorts entries by search score
const SortScore = SortOrder(0)

// SortRecent sorts entries by descending modified date
const SortRecent = SortOrder(1)

// SortName sorts entries alphabetically by name
const SortName = SortOrder(2)

// The data variable stores all the things that get saved.
var data = root{
	Names: make(map[string]model.Entry),
}

func (r *root) lock() {
	r.mux.Lock()
}
func (r *root) unlock() {
	r.mux.Unlock()
}

// PutEntry adds or replaces the given entry in the collection.
func PutEntry(entry model.Entry) error {
	data.lock()
	slug := entry.Slug()
	data.Names[slug] = entry
	data.unlock()
	return IndexEntry(entry)
}

// DeleteEntry removes the specified entry from the collection.
func DeleteEntry(slug string) (bool, error) {
	data.lock()
	defer data.unlock()
	_, exists := GetEntryFromIndex(slug)
	if !exists {
		return false, nil
	}
	if _, exists := data.Names[slug]; exists {
		delete(data.Names, slug)
	}
	deletes = append(deletes, slug)
	return true, RemoveFromIndex(slug)
}

// GetEntryFromStorage returns a single entry suitable for editing or throws an error.
func GetEntryFromStorage(slug string) (model.Entry, error) {
	// check for modified and unsaved entry first
	pendingSave, exists := data.Names[slug]
	if exists {
		return pendingSave, nil
	}
	// read entry content from storage
	entry, err := persist.ReadEntry(slug)
	if err != nil {
		return model.Entry{}, err
	}
	return entry, nil
}

// Init reads data stored on the file system and initializes application variables.
// homeDir provides an optional override to the default location of ~/.memory where
// settings and data are stored.
func Init(homeDir string) error {
	if inited {
		return nil
	}
	// set home dir
	if homeDir == "" {
		homeDir = util.GetHomeDir() + string(os.PathSeparator) + config.MemoryHome
	}
	config.MemoryHome = homeDir
	if err := localfs.InitHome(); err != nil {
		return err
	}
	// load config
	if util.PathExists(config.SettingsPath()) {
		settings := config.StoredSettings{}
		if err := localfs.Load(config.SettingsPath(), &settings); err != nil {
			return fmt.Errorf("failed to load settings: %s", err.Error())
		}
		config.UpdateSettingsFromStorage(settings)
		// initialize settings file
	} else if err := localfs.Save(config.SettingsPath(), config.GetSettingsForStorage()); err != nil {
		return fmt.Errorf("failed to initialize settings: %w", err)
	}
	// load data
	if err := initSearch(); err != nil {
		return err
	}
	inited = true
	return nil
}

// Shutdown performs cleanup prior to exiting application
func Shutdown() error {
	return closeSearch()
}

// RenameEntry changes an entry name and updates associated data structures.
func RenameEntry(name string, newName string) error {
	data.lock()
	slug := util.GetSlug(name)
	newSlug := util.GetSlug(newName)
	defer data.unlock()
	_, exists := GetEntryFromIndex(newSlug)
	if exists {
		return fmt.Errorf("an entry named %s (or very similar) already exists", newName)
	}
	entry, err := GetEntryFromStorage(slug)
	if err != nil {
		return err
	}
	if _, err = DeleteEntry(slug); err != nil {
		return err
	}
	entry.Name = newName
	if err = PutEntry(entry); err != nil {
		return err
	}
	return nil
}

// Save writes application data to file storage.
func Save() error {
	data.lock()
	defer data.unlock()
	for slug, entry := range data.Names {
		if err := persist.SaveEntry(entry); err != nil {
			return fmt.Errorf("failed to save %s: %s", slug, err.Error())
		}
	}
	data.Names = make(map[string]model.Entry)
	for _, slug := range deletes {
		err := persist.DeleteEntry(slug)
		if err != nil {
			return fmt.Errorf("failed to delete %s: %s", slug, err.Error())
		}
	}
	deletes = []string{}
	return nil
}

// filterType returns true if the entry is one of the true EntryTypes
func filterType(entry model.Entry, types model.EntryTypes) bool {
	if types.HasAll() {
		return true
	}
	switch entry.Type {
	case model.EntryTypeEvent:
		return types.Event
	case model.EntryTypeNote:
		return types.Note
	case model.EntryTypePerson:
		return types.Person
	case model.EntryTypePlace:
		return types.Place
	case model.EntryTypeThing:
		return types.Thing
	}
	return false
}

func sortEntries(arr []model.Entry, field string, ascending bool) {
	var less func(i, j int) bool
	switch field {
	case "Modified":
		less = func(i, j int) bool {
			if ascending {
				return arr[i].Modified.UnixNano() < arr[j].Modified.UnixNano()
			}
			return arr[i].Modified.UnixNano() > arr[j].Modified.UnixNano()
		}
	default: // Name
		less = func(i, j int) bool {
			if ascending {
				return arr[i].Name < arr[j].Name
			}
			return arr[i].Name > arr[j].Name
		}
	}
	sort.Slice(arr, less)
}

// GetSortedNames returns a slice of all entry names sorted alphabetically.
func GetSortedNames() ([]string, error) {
	keys, err := IndexedSlugs()
	if err != nil {
		return []string{}, err
	}
	sort.Strings(keys)
	return keys, nil
}
