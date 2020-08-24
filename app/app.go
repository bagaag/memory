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
	"errors"
	"fmt"
	"memory/app/config"
	"memory/app/persist"
	"memory/util"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/gosimple/slug"
)

var inited = false // run Init() only once

var deletes = []string{}

// root contains all the data to be saved
type root struct {
	Names map[string]Entry
	mux   sync.Mutex
}

// EntryResults is used to contain the results of GetEntries and the settings used
// to generate those results.
type EntryResults struct {
	Entries  []Entry
	Types    EntryTypes
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
	Names: make(map[string]Entry),
}

func (r *root) lock() {
	r.mux.Lock()
}
func (r *root) unlock() {
	r.mux.Unlock()
}

// GetSlug converts a string into a slug
func GetSlug(s string) string {
	return slug.Make(s)
}

// PutEntry adds or replaces the given entry in the collection.
func PutEntry(entry Entry) {
	data.lock()
	slug := entry.Slug()
	data.Names[slug] = entry
	data.unlock()
	IndexEntry(entry)
}

// DeleteEntry removes the specified entry from the collection.
func DeleteEntry(slug string) bool {
	data.lock()
	defer data.unlock()
	_, exists := GetEntryFromIndex(slug)
	if !exists {
		return false
	}
	if _, exists := data.Names[slug]; exists {
		delete(data.Names, slug)
	}
	deletes = append(deletes, slug)
	RemoveFromIndex(slug)
	return true
}

// HasAll returns true if either all are true or all are false.
func (t EntryTypes) HasAll() bool {
	if (t.Note && t.Event && t.Person && t.Place && t.Thing) ||
		(!t.Note && !t.Event && !t.Person && !t.Place && !t.Thing) {
		return true
	}
	return false
}

// String returns a string representation of the selected types.
func (t EntryTypes) String() string {
	s := "All types"
	if !t.HasAll() {
		a := []string{}
		if t.Note {
			// TODO: Codify plural entry types in entry.go
			a = append(a, "Notes")
		}
		if t.Event {
			a = append(a, "Events")
		}
		if t.Person {
			a = append(a, "People")
		}
		if t.Place {
			a = append(a, "Places")
		}
		if t.Thing {
			a = append(a, "Things")
		}
		s = strings.Join(a, ", ")
	}
	return s
}

// GetEntryFromStorage returns a single entry suitable for editing or throws an error.
func GetEntryFromStorage(slug string) (Entry, bool, error) {
	// check for modified and unsaved entry first
	pendingSave, exists := data.Names[slug]
	if exists {
		return pendingSave, true, nil
	}
	// make sure entry exists in storage
	if !persist.EntryExists(slug) {
		return Entry{}, false, nil
	}
	// read entry content from storage
	content, modified, err := persist.ReadEntry(slug)
	if err != nil {
		return Entry{}, false, err
	}
	// parse entry content into Entry
	entry, err := ParseYamlDown(content)
	if err != nil {
		return Entry{}, true, err
	}
	entry.Modified = modified
	//TODO: remove this and implement caching if it seems to happen too often
	fmt.Println("Read", slug, "from storage")
	return entry, true, nil
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
	persist.InitHome()
	// load config
	if persist.PathExists(config.SettingsPath()) {
		settings := config.StoredSettings{}
		if err := persist.Load(config.SettingsPath(), &settings); err != nil {
			return fmt.Errorf("failed to load settings: %s", err.Error())
		}
		config.UpdateSettingsFromStorage(settings)
		// initialize settings file
	} else if err := persist.Save(config.SettingsPath(), config.GetSettingsForStorage()); err != nil {
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
	slug := GetSlug(name)
	newSlug := GetSlug(newName)
	defer data.unlock()
	_, exists := GetEntryFromIndex(newSlug)
	if exists {
		return fmt.Errorf("an entry named %s (or very similar) already exists", newName)
	}
	entry, exists, err := GetEntryFromStorage(slug)
	if !exists {
		return fmt.Errorf("an entry named %s does not exist", name)
	}
	if err != nil {
		return err
	}
	DeleteEntry(slug)
	entry.Name = newName
	PutEntry(entry)
	return nil
}

// Save writes application data to file storage.
func Save() error {
	data.lock()
	defer data.unlock()
	for slug, entry := range data.Names {
		content, err := RenderYamlDown(entry)
		if err != nil {
			return fmt.Errorf("failed to render %s: %s", slug, err.Error())
		}
		err = persist.SaveEntry(slug, content)
		if err != nil {
			return fmt.Errorf("failed to save %s: %s", slug, err.Error())
		}
	}
	data.Names = make(map[string]Entry)
	for _, slug := range deletes {
		err := persist.DeleteEntry(slug)
		if err != nil {
			return fmt.Errorf("failed to delete %s: %s", slug, err.Error())
		}
	}
	deletes = []string{}
	return nil
}

// ValidateEntryName returns an error if the given name is invalid.
func ValidateEntryName(name string) error {
	if len(name) == 0 {
		return errors.New("name cannot be an empty string")
	}
	if strings.HasPrefix(name, " ") {
		return errors.New("name cannot start with a space")
	}
	if strings.HasSuffix(name, " ") {
		return errors.New("name cannot end with a space")
	}
	if strings.Contains(name, "\n") || strings.Contains(name, "\r") {
		return errors.New("name cannot contain line breaks")
	}
	if strings.Contains(name, "\t") {
		return errors.New("name cannot contain tab characters")
	}
	if strings.Contains(name, "  ") {
		return errors.New("name cannot more than 1 consecutive space")
	}
	if strings.HasPrefix(name, "!") {
		return errors.New("name cannot start with a ! character")
	}
	if strings.Contains(name, "[") || strings.Contains(name, "]") {
		return errors.New("name cannot contain [ or ]")
	}
	if len(name) > config.MaxNameLen {
		return fmt.Errorf("name length cannot exceed %d", config.MaxNameLen)
	}
	return nil
}

// filterType returns true if the entry is one of the true EntryTypes
func filterType(entry Entry, types EntryTypes) bool {
	if types.HasAll() {
		return true
	}
	switch entry.Type {
	case EntryTypeEvent:
		return types.Event
	case EntryTypeNote:
		return types.Note
	case EntryTypePerson:
		return types.Person
	case EntryTypePlace:
		return types.Place
	case EntryTypeThing:
		return types.Thing
	}
	return false
}

// searchMatches returns a score from 0 to 1 if the entry matches, where 0 is
// a miss and 1 is a perfect match on entry name.
func searchMatches(keywords string, entry Entry) {
	keywords = strings.ToLower(keywords)
	//if strings.ToLower(entry.Name) == keyw
}

func sortEntries(arr []Entry, field string, ascending bool) {
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
