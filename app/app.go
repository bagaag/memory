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
	"memory/app/search"
	"memory/impl"
	"memory/util"
	"regexp"
	"sort"
	"strings"
)

type Memory struct {
	Persist persist.Persister // stores Entries
	Search  search.Searcher   // provides Entry search
	linkExp *regexp.Regexp    // initialized on first use
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
	// load data
	// TODO: use config to determine which DI implementations to use
	m := Memory{}
	persistConfig := impl.SimplePersistConfig{
		EntryPath: config.EntriesPath(),
		FilePath:  config.FilesPath(),
		EntryExt:  config.EntryExt,
	}
	persister, err := impl.NewSimplePersist(persistConfig)
	if err != nil {
		return nil, err
	} else {
		m.Persist = &persister
	}
	// load search provider
	searchConfig := impl.BleveSearchConfig{
		IndexDir:  config.SearchPath(),
		Persister: &persister,
	}
	searcher, err := impl.NewBleveSearch(searchConfig)
	if err != nil {
		return nil, err
	} else {
		m.Search = &searcher
	}
	return &m, nil
}

// PutEntry adds or replaces the given entry in the collection.
func (m *Memory) PutEntry(entry model.Entry) error {
	if err := m.Persist.SaveEntry(entry); err != nil {
		return err
	}
	return m.Search.IndexEntry(entry)
}

// DeleteEntry removes the specified entry from the collection.
func (m *Memory) DeleteEntry(slug string) error {
	_, err := m.Search.GetEntry(slug)
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
	exists, _ := m.EntryExists(newSlug)
	if exists {
		return model.Entry{}, fmt.Errorf("an entry named %s (or very similar) already exists", newName)
	}
	if err := m.Search.RemoveFromIndex(oldSlug); err != nil {
		return model.Entry{}, err
	}
	var err error
	var entry model.Entry
	if entry, err = m.Persist.RenameEntry(oldName, newName); err != nil {
		return entry, err
	}
	return entry, nil
}

// ParseLinks looks for [Name] links within the given string and
// returns a slice of index pairs found. Links that cannot be
// resolved are replaced with a ! prefix in the parsed return
// value, as in [!Not Found].
func (m *Memory) ParseLinks(s string) (string, []string) {
	// init return values
	parsed := s
	links := []string{}
	// compile links regexp
	if m.linkExp == nil {
		var err error
		m.linkExp, err = regexp.Compile("\\[([[:alnum:]?][^~\\]]*)\\]\\(?")
		if err != nil {
			fmt.Println("Error compiling link regexp:", err)
			return s, []string{}
		}
	}
	// get [links]
	results := m.linkExp.FindAllStringIndex(s, -1)
	for _, pair := range results {
		link := s[pair[0]:pair[1]]
		// ignore external links, which are followed immediately by "("
		if strings.HasSuffix(link, "(") {
			continue
		}
		// strip off brackets, remove line breaks and consecutive spaces
		name := link[1 : len(link)-1]
		name = strings.ReplaceAll(name, "\n", " ")
		for strings.Contains(name, "  ") {
			name = strings.ReplaceAll(name, "  ", " ") //TODO: use regex to replace 2+ whitespace
		}
		// remove ! if it's already there (! prefix indicates non-existent entry)
		hadBang := false
		if strings.HasPrefix(name, "?") {
			name = name[1:]
			hadBang = true
		}
		slug := util.GetSlug(name)
		// add to results if exists, otherwise add ! prefix
		if exists, _ := m.EntryExists(slug); exists {
			// remove erroneous ! prefix if needed
			if hadBang {
				linkWithoutBang := "[" + link[2:]
				parsed = strings.Replace(parsed, link, linkWithoutBang, 1)
			}
		} else if !hadBang {
			// entry doesn't exist, add a ! if needed
			link404 := "[?" + link[1:]
			parsed = strings.Replace(parsed, link, link404, 1)
		}
		if !util.StringSliceContains(links, name) {
			links = append(links, slug)
		}
	}
	return parsed, links
}

// ResolveLinks accepts a slice of Entry names and returns
// a slice of Entries that exist with those names.
func (m *Memory) ResolveLinks(links []string) []model.Entry {
	resolved := []model.Entry{}
	for _, slug := range links {
		if entry, err := m.Search.GetEntry(slug); err != nil {
			resolved = append(resolved, entry)
		}
	}
	return resolved
}

// UpdateLinks populates the LinksTo and LinkedFrom slices on all entries by
// parsing the descriptions for links.
func (m *Memory) UpdateLinks() error {
	fromLinks := make(map[string][]string)
	slugs, err := m.Search.IndexedSlugs()
	if err != nil {
		return err
	}
	for _, slug := range slugs {
		// parse and save outgoing links for this entry
		entry, err := m.Search.GetEntry(slug)
		if err != nil {
			searchText := entry.Description
			newDesc, links := m.ParseLinks(searchText)
			entry.Description = newDesc
			entry.LinksTo = links
			entry.LinkedFrom = []string{}
			if err := m.Search.IndexEntry(entry); err != nil {
				return err
			}
			fromSlug := entry.Slug()
			// add links in reverse direction
			for _, toSlug := range links {
				slugs, exists := fromLinks[toSlug]
				if !exists {
					slugs = []string{fromSlug}
				} else if !util.StringSliceContains(slugs, fromSlug) {
					slugs = append(slugs, fromSlug)
				}
				fromLinks[toSlug] = slugs
			}
		}
	}
	// save the fromLinks in corresponding entries
	for slug, linkedFrom := range fromLinks {
		entry, err := m.Search.GetEntry(slug)
		if err != nil {
			entry.LinkedFrom = linkedFrom
			if err := m.Search.IndexEntry(entry); err != nil {
				return err
			}
		}
	}
	return nil
}

// BrokenLinks returns a map of string slices containing names of linked-to pages that don't
// exist; the name of the page containing the link is the key.
func (m *Memory) BrokenLinks() (map[string][]string, error) {
	ret := make(map[string][]string)
	slugs, err := m.Search.IndexedSlugs()
	if err != nil {
		return ret, err
	}
	for _, slug := range slugs {
		fromEntry, _ := m.Search.GetEntry(slug)
		for _, toName := range fromEntry.LinksTo {
			if !util.StringSliceContains(slugs, toName) {
				var brokenLinks []string
				var existingList bool
				if brokenLinks, existingList = ret[fromEntry.Name]; existingList {
					brokenLinks = append(brokenLinks, toName)
					sort.Strings(brokenLinks)
				} else {
					brokenLinks = []string{toName}
				}
				ret[fromEntry.Name] = brokenLinks
			}
		}
	}
	return ret, nil
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
		entry, _ := m.Search.GetEntry(slug)
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
	if entry, err := m.Search.GetEntry(slug); err != nil {
		return "", err
	} else {
		return entry.Name, nil
	}
}

// EntryExists is a shortcut to calling GetEntry and testing the resulting error against EntryNotFound
func (m *Memory) EntryExists(slug string) (bool, error) {
	_, err := m.GetEntry(slug)
	if err != nil {
		if _, notFound := err.(model.EntryNotFound); notFound {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
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
