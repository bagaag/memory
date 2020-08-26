/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package model

import (
	"errors"
	"fmt"
	"memory/app/config"
	"memory/util"
	"strings"
	"time"
)

// Entry represents a Person, Place, Thing, Event or Note.
type Entry struct {
	Name        string
	Description string
	Tags        []string
	LinksTo     []string
	LinkedFrom  []string
	Created     time.Time
	Modified    time.Time
	Type        EntryType `json:"EntryType"`
	Start       string    // Events
	End         string    // Events
	Latitude    string    // Place
	Longitude   string    // Place
	Address     string    // Place
	Custom      map[string]string
	Exclude     bool // Supports ability to search for all entries
}

// Slug returns the slug for this entry.
//TODO: Replace instances of GetSlug(entry.Name)
func (entry *Entry) Slug() string {
	return util.GetSlug(entry.Name)
}

// NewEntry initializes and returns an Entry object.
func NewEntry(entryType EntryType, name string, description string, tags []string) Entry {
	now := time.Now()
	entry := Entry{
		Name:        name,
		Description: description,
		Tags:        tags,
		LinksTo:     []string{},
		LinkedFrom:  []string{},
		Modified:    now,
		Type:        entryType,
		Custom:      make(map[string]string),
	}
	return entry
}

// EntryTypes is used to indicate one or more entry types in a single argument
type EntryTypes struct {
	Note   bool
	Event  bool
	Person bool
	Place  bool
	Thing  bool
}

// EntryType is an 'enum' of entry types.
type EntryType = string

// EntryTypeNote indicates a note.
const EntryTypeNote = "Note"

// EntryTypeEvent indicates an event.
const EntryTypeEvent = "Event"

// EntryTypePerson indicates a person.
const EntryTypePerson = "Person"

// EntryTypePlace indicates a place.
const EntryTypePlace = "Place"

// EntryTypeThing indicates a thing.
const EntryTypeThing = "Thing"

// TagsString returns the entry's tags as a comma-separated string.
func (entry Entry) TagsString() string {
	return strings.Join(entry.Tags, ",")
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

// BleveType implements the alternate bleve.Classifier interface to avoid a
// naming conflict with .Type.
func (entry *Entry) BleveType() string {
	return "Entry"
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
