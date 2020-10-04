/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

//Package app contains an API for interacting with the application
//that is not bound to a particular UI.
package memory

import (
	"fmt"
	"memory/app/attachment"
	"memory/app/config"
	"memory/app/localfs"
	"memory/app/model"
	"memory/app/persist"
	"memory/app/search"
	"memory/util"
	"sort"
)

type Memory struct {
	Persist persist.Persister   // provides Entry storage
	Search  search.Searcher     // provides Entry search
	Attach  attachment.Attacher // provides Attachment storage
}

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
	// load data provider
	m := Memory{}
	persistConfig := persist.SimplePersistConfig{
		EntryPath: config.EntriesPath(),
		FilePath:  config.FilesPath(),
	}
	persister, err := persist.NewSimplePersist(persistConfig)
	if err != nil {
		return nil, err
	} else {
		m.Persist = &persister
	}
	// load search provider
	searchConfig := search.BleveSearchConfig{
		IndexDir:  config.SearchPath(),
		Persister: &persister,
	}
	searcher, err := search.NewBleveSearch(searchConfig)
	if err != nil {
		return nil, err
	} else {
		m.Search = &searcher
	}
	// load attachment provider
	attacher := attachment.LocalAttachmentStore{StoragePath: config.FilesPath()}
	m.Attach = &attacher
	return &m, nil
}

// PutEntry adds or replaces the given entry in the collection.
func (m *Memory) PutEntry(entry model.Entry) error {
	if m.EntryExists(entry.Slug()) {
		if existing, err := m.GetEntry(entry.Slug()); err == nil {
			entry.Created = existing.Created
		}
	}
	if err := m.Persist.SaveEntry(entry); err != nil {
		return err
	}
	return m.Search.IndexEntry(entry)
}

// DeleteEntry removes the specified entry from the collection.
func (m *Memory) DeleteEntry(slug string) error {
	_, err := m.Search.Stub(slug)
	if err != nil {
		return err
	}
	if err := m.Persist.DeleteEntry(slug); err != nil {
		return err
	}
	return m.Search.RemoveFromIndex(slug)
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
	// check entry existence
	if m.EntryExists(newSlug) {
		return model.Entry{}, fmt.Errorf("an entry named %s (or very similar) already exists", newName)
	}
	// remove from search
	if err := m.Search.RemoveFromIndex(oldSlug); err != nil {
		return model.Entry{}, err
	}
	// update entry persistence
	var err error
	var entry model.Entry
	if entry, err = m.Persist.RenameEntry(oldName, newName); err != nil {
		return entry, err
	}
	// update attachment persistence
	if err = m.Attach.RenameEntry(oldSlug, newSlug); err != nil {
		return entry, err
	}
	// update search index
	if err = m.Search.IndexEntry(entry); err != nil {
		return entry, err
	}
	return entry, nil
}

// GetTags returns a map of all defined tags, each with a sorted slice of
// associated entry names.
func (m *Memory) GetTags() (map[string][]string, error) {
	tags := make(map[string][]string)
	slugs, err := m.Search.IndexedSlugs()
	if err != nil {
		return tags, err
	}
	for _, slug := range slugs {
		entry, _ := m.Search.Stub(slug)
		for _, tag := range entry.Tags {
			names, exists := tags[tag]
			if !exists {
				names = []string{entry.Name}
			} else {
				if !util.StringSliceContains(names, entry.Name) {
					names = append(names, entry.Name)
					sort.Strings(names)
				}
			}
			tags[tag] = names
		}
	}
	return tags, nil
}

// GetSortedTags takes the output of GetTags and returns a sorted
// slice of tags.
func (m *Memory) GetSortedTags(tags map[string][]string) []string {
	keys := []string{}
	for tag := range tags {
		keys = append(keys, tag)
	}
	sort.Strings(keys)
	return keys
}

// NameFromSlug swaps a slug with an Entry name.
func (m *Memory) NameFromSlug(slug string) (string, error) {
	if entry, err := m.Search.Stub(slug); err != nil {
		return "", err
	} else {
		return entry.Name, nil
	}
}

// EntryExists is a shortcut to calling GetEntry and testing the resulting error against EntryNotFound
func (m *Memory) EntryExists(slug string) bool {
	return m.Persist.EntryExists(slug)
}

// tagMatches returns true if any of the tags in searchTags match the tags
// on the provided Entry.
func (m *Memory) tagMatches(entry model.Entry, searchTags []string, matchesAll bool) bool {
	for _, searchTag := range searchTags {
		matches := util.StringSliceContains(entry.Tags, searchTag)
		if matches && !matchesAll {
			return true
		} else if !matches && matchesAll {
			return false
		}
	}
	return matchesAll
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
