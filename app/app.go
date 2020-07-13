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
	"memory/app/config"
	"memory/app/model"
	"memory/app/persist"
	"sort"
	"strings"
)

const dataVersion = 1

// root contains all the data to be saved
type root struct {
	Names map[string]model.Entry
}

// saveData is just the data we need to save to file - eliminating any
// calculated data in the root struct.
type saveData struct {
	Names   map[string]model.Entry
	Version int // for migrations
}

// EntryTypes is used to indicate one or more entry types in a single argument
type EntryTypes struct {
	Note   bool
	Event  bool
	Person bool
	Place  bool
	Thing  bool
}

// EntryResults is used to contain the results of GetEntries and the settings used
// to generate those results.
type EntryResults struct {
	Entries    []model.Entry
	Types      EntryTypes
	StartsWith string
	Contains   string
	Search     string
	Tags       []string
	Sort       SortOrder
	Limit      int
}

// PutEntry adds or replaces the given entry in the collection.
func PutEntry(entry model.Entry) {
	data.Names[entry.Name()] = entry
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
var data = root{}

// Init reads data stored on the file system
// and initializes application variable.
func Init() error {
	if persist.PathExists(config.SavePath()) {
		fromsave := saveData{}
		fromsave.Names = make(map[string]model.Entry)
		if err := persist.Load(config.SavePath(), &fromsave); err != nil {
			return err
		}
		//TODO: handle version difference w/ migration
		data.Names = fromsave.Names
	}
	return nil
}

// filterType returns true if the entry is one of the true EntryTypes
func filterType(entry model.Entry, types EntryTypes) bool {
	if types.HasAll() {
		return true
	}
	switch entry.(type) {
	case model.Note:
		return types.Note
	// TODO: add Event, Person, Place, Thing models as they're created here
	default:
		return false
	}
}

// GetEntries returns an array of entries of the specified type(s) with
// specified filters and sorting applied.
func GetEntries(types EntryTypes, startsWith string, contains string,
	search string, tags []string, sort SortOrder, limit int) EntryResults {

	// holds the results
	entries := []model.Entry{}

	// convert filters to lower case
	startsWith = strings.ToLower(startsWith)
	contains = strings.ToLower(startsWith)
	for i, tag := range tags {
		tags[i] = strings.ToLower(tag)
	}

	// run through the collection and apply filters
	for _, entry := range data.Names {
		lowerName := strings.ToLower(entry.Name())

		if !filterType(entry, types) {
			continue
		}
		if startsWith != "" && !strings.HasPrefix(lowerName, startsWith) {
			continue
		}
		if contains != "" && !strings.Contains(lowerName, contains) {
			continue
		}
		//TODO: implement search
		if len(tags) > 0 && !tagMatches(entry, tags) {
			continue
		}
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
		limit = 999
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

// EntryCount returns the total number of entries under management.
func EntryCount() int {
	return len(data.Names)
}

// GetEntry returns a single entry or throws an error.
func GetEntry(entryName string) (model.Entry, bool) {
	entry, exists := data.Names[entryName]
	return entry, exists
}

// Save writes application data to file storage.
func Save() error {
	tosave := saveData{Names: data.Names, Version: dataVersion}
	return persist.Save(config.SavePath(), tosave)
}

// tagMatches returns true if any of the tags in searchTags match the tags
// on the provided Entry.
func tagMatches(entry model.Entry, searchTags []string) bool {
	for _, searchTag := range searchTags {
		for _, tag := range entry.Tags() {
			if tag == searchTag {
				return true
			}
		}
	}
	return false
}

func sortEntries(arr []model.Entry, field string, ascending bool) {
	var less func(i, j int) bool
	switch field {
	case "Modified":
		less = func(i, j int) bool {
			if ascending {
				return arr[i].Modified().UnixNano() < arr[j].Modified().UnixNano()
			}
			return arr[i].Modified().UnixNano() > arr[j].Modified().UnixNano()
		}
	default: // Name
		less = func(i, j int) bool {
			if ascending {
				return arr[i].Name() < arr[j].Name()
			}
			return arr[i].Name() > arr[j].Name()
		}
	}
	sort.Slice(arr, less)
}
