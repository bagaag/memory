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
)

const dataVersion = 1

var inited = false // run Init() only once

// root contains all the data to be saved
type root struct {
	Names map[string]Entry
}

// saveData is just the data we need to save to file - eliminating any
// calculated data in the root struct.
type saveData struct {
	Entries []Entry
	Version int // for migrations
}

// EntryResults is used to contain the results of GetEntries and the settings used
// to generate those results.
type EntryResults struct {
	Entries    []Entry
	Types      EntryTypes
	StartsWith string
	Contains   string
	Search     string
	Tags       []string
	Sort       SortOrder
	Limit      int
}

// PutEntry adds or replaces the given entry in the collection.
func PutEntry(entry Entry) {
	data.Names[entry.Name] = entry
}

// DeleteEntry removes the specified entry from the collection.
func DeleteEntry(name string) bool {
	_, exists := data.Names[name]
	if !exists {
		return false
	}
	delete(data.Names, name)
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

// SortOrder is used to indicate one of the Sort constants
type SortOrder int

// SortRecent sorts entries by descending modified date
const SortRecent = SortOrder(0)

// SortName sorts entries alphabetically by name
const SortName = SortOrder(1)

// The data variable stores all the things that get saved.
var data = root{
	Names: make(map[string]Entry),
}

// EntryCount returns the total number of entries under management.
func EntryCount() int {
	return len(data.Names)
}

// GetEntries returns an array of entries of the specified type(s) with
// specified filters and sorting applied.
func GetEntries(types EntryTypes, startsWith string, contains string,
	search string, tags []string, sort SortOrder, limit int) EntryResults {

	// holds the results
	entries := []Entry{}

	// convert filters to lower case
	startsWith = strings.ToLower(startsWith)
	contains = strings.ToLower(startsWith)
	for i, tag := range tags {
		tags[i] = strings.ToLower(tag)
	}

	// run through the collection and apply filters
	for _, entry := range data.Names {
		lowerName := strings.ToLower(entry.Name)

		if !filterType(entry, types) {
			continue
		}
		if startsWith != "" && !strings.HasPrefix(lowerName, startsWith) {
			continue
		}
		if contains != "" && !strings.Contains(lowerName, contains) {
			continue
		}
		if len(tags) > 0 && !tagMatches(entry, tags) {
			continue
		}
		//TODO: implement search
		// if we made it this far, add to return slice
		entries = append(entries, entry)
	}

	// sort entries
	if sort == SortName {
		sortEntries(entries, "Name", true)
	} else { // SortRecent
		sortEntries(entries, "Modified", false)
	}
	// limit sorted results
	if limit <= 0 {
		limit = 2147483647 // int32 max just to be safe
	}
	if len(entries) > limit {
		entries = entries[:limit]
	}
	return EntryResults{
		Entries:    entries,
		Types:      types,
		StartsWith: startsWith,
		Contains:   contains,
		Search:     search,
		Tags:       tags,
		Sort:       sort,
		Limit:      limit,
	}
}

// RefreshResults re-runs a search query and gets fresh results to avoid showing
// stale entries when results are revisited.
func RefreshResults(results EntryResults) EntryResults {
	return GetEntries(results.Types, results.StartsWith, results.Contains,
		results.Search, results.Tags, results.Sort, results.Limit)
}

// GetEntry returns a single entry or throws an error.
func GetEntry(entryName string) (Entry, bool) {
	entry, exists := data.Names[entryName]
	return entry, exists
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
	// load config
	if persist.PathExists(config.SettingsPath()) {
		settings := config.StoredSettings{}
		if err := persist.Load(config.SettingsPath(), &settings); err != nil {
			return fmt.Errorf("failed to load settings: %s", err.Error)
		}
		config.UpdateSettingsFromStorage(settings)
		// initialize settings file
	} else if err := persist.Save(config.SettingsPath(), config.GetSettingsForStorage()); err != nil {
		return fmt.Errorf("failed to initialize settings: %w", err)
	}
	// load data
	if persist.PathExists(config.SavePath()) {
		// read saved file into saveData struct
		fromSave := saveData{}
		if err := persist.Load(config.SavePath(), &fromSave); err != nil {
			return err
		}
		//TODO: handle version difference w/ migration
		// setup runtime data structure from saved data
		for _, entry := range fromSave.Entries {
			data.Names[entry.Name] = entry
		}
		PopulateLinks()
	}
	inited = true
	return nil
}

// RenameEntry changes an entry name and updates associated data structures.
func RenameEntry(name string, newName string) error {
	_, exists := GetEntry(newName)
	if exists {
		return fmt.Errorf("an entry named %s already exists", newName)
	}
	entry, exists := GetEntry(name)
	if !exists {
		return fmt.Errorf("an entry named %s does not exist", name)
	}
	DeleteEntry(name)
	entry.Name = newName
	PutEntry(entry)
	return nil
}

// Save writes application data to file storage.
func Save() error {
	toSave := saveData{Version: dataVersion}
	toSave.Entries = []Entry{}
	for _, entry := range data.Names {
		toSave.Entries = append(toSave.Entries, entry)
	}
	return persist.Save(config.SavePath(), toSave)
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

// tagMatches returns true if any of the tags in searchTags match the tags
// on the provided Entry.
func tagMatches(entry Entry, searchTags []string) bool {
	for _, searchTag := range searchTags {
		if util.StringSliceContains(entry.Tags, searchTag) {
			return true
		}
	}
	return false
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
