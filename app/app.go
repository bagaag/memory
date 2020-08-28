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
	"memory/impl"
	"memory/util"
	"sort"
)

type Memory struct {
	Persist persist.Persister // stores Entries
	//Search *search.Search // provides Entry search
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

// Init reads data stored on the file system and initializes application variables.
// homeDir provides an optional override to the default location of ~/.memory where
// settings and local data are stored. Pass "" for homeDir to use config value.
func Init(homeDir string) (*Memory, error) {
	// allow for optional override of default home location
	if homeDir != "" {
		config.MemoryHome = homeDir
	} else {
		config.MemoryHome = util.GetHomeDir() + localfs.Slash + config.MemoryHome
	}
	if err := localfs.InitHome(); err != nil {
		return nil, err
	}
	// load config
	// TODO: use DI for config & replace w/ https://github.com/uber-go/config
	if localfs.PathExists(config.SettingsPath()) {
		settings := config.StoredSettings{}
		if err := localfs.Load(config.SettingsPath(), &settings); err != nil {
			return nil, fmt.Errorf("failed to load settings: %s", err.Error())
		}
		config.UpdateSettingsFromStorage(settings)
		// initialize settings file
	} else if err := localfs.Save(config.SettingsPath(), config.GetSettingsForStorage()); err != nil {
		return nil, fmt.Errorf("failed to initialize settings: %w", err)
	}
	// load data
	// TODO: use config to determine which DI implementations to use
	m := Memory{}
	simpleCfg := impl.SimplePersistConfig{
		EntryPath: config.EntriesPath(),
		FilePath:  config.FilesPath(),
		EntryExt:  config.EntryExt,
	}
	if sp, err := impl.NewSimplePersist(simpleCfg); err != nil {
		return nil, err
	} else {
		m.Persist = &sp
	}
	// TODO: use DI for search
	if err := initSearch(); err != nil {
		return nil, err
	}
	return &m, nil
}

// PutEntry adds or replaces the given entry in the collection.
func (m *Memory) PutEntry(entry model.Entry) error {
	if err := m.Persist.SaveEntry(entry); err != nil {
		return err
	}
	return IndexEntry(entry)
}

// DeleteEntry removes the specified entry from the collection.
func (m *Memory) DeleteEntry(slug string) error {
	_, exists := GetEntryFromIndex(slug)
	if !exists {
		return persist.EntryNotFound{Slug: slug}
	}
	if err := m.Persist.DeleteEntry(slug); err != nil {
		return err
	}
	return RemoveFromIndex(slug)
}

// GetEntryFromStorage returns a single entry suitable for editing or throws an error.
func (m *Memory) GetEntry(slug string) (model.Entry, error) {
	return m.Persist.ReadEntry(slug)
}

// RenameEntry changes an entry name and updates associated data structures, returning
// the slug for the renamed entry.
func (m *Memory) RenameEntry(oldName string, newName string) (model.Entry, error) {
	oldSlug := util.GetSlug(oldName)
	newSlug := util.GetSlug(newName)
	_, exists := GetEntryFromIndex(newSlug)
	if exists {
		return model.Entry{}, fmt.Errorf("an entry named %s (or very similar) already exists", newName)
	}
	if err := RemoveFromIndex(oldSlug); err != nil {
		return model.Entry{}, err
	}
	var err error
	var entry model.Entry
	if entry, err = m.Persist.RenameEntry(oldName, newName); err != nil {
		return entry, err
	}
	return entry, nil
}

// TODO: move to simple search impl
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

// TODO: move to simple search impl
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

// TODO: move to simple search impl
// GetSortedNames returns a slice of all entry names sorted alphabetically.
func GetSortedNames() ([]string, error) {
	keys, err := IndexedSlugs()
	if err != nil {
		return []string{}, err
	}
	sort.Strings(keys)
	return keys, nil
}
