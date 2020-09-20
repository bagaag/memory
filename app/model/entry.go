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
	Created     time.Time
	Modified    time.Time
	Type        EntryType `json:"EntryType"`
	Start       FlexDate  // Events
	End         FlexDate  // Events
	Latitude    string    // Place
	Longitude   string    // Place
	Address     string    // Place
	Custom      map[string]string
	Files       []File
	populated   bool // Indicates that full details are populated
}

// Slug returns the slug for this entry.
//TODO: Replace instances of GetSlug(entry.Name)
func (entry *Entry) Slug() string {
	return util.GetSlug(entry.Name)
}

// Populated indicates whether full details are populated.
func (entry *Entry) Populated() bool {
	return entry.populated
}

// SetPopulated indicates whether full details are populated.
func (entry *Entry) SetPopulated(p bool) {
	entry.populated = p
}

// NewEntry initializes and returns an Entry object.
func NewEntry(entryType EntryType, name string, description string, tags []string) Entry {
	now := time.Now()
	entry := Entry{
		Name:        name,
		Description: description,
		Tags:        tags,
		Modified:    now,
		Type:        entryType,
		Custom:      make(map[string]string),
		Files:       []File{},
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

const EntryTypeNote = "Note"
const EntryTypeEvent = "Event"
const EntryTypePerson = "Person"
const EntryTypePlace = "Place"
const EntryTypeThing = "Thing"

// Precision is an 'enum' of int values
type Precision = int

const PrecisionNone = -1
const PrecisionYear = 0
const PrecisionMonth = 1
const PrecisionDay = 2

// FlexDate is a string in the form of 2006, 2006-01 or 2006-01-02
type FlexDate = string

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

// EntryNotFound is a custom error type to indicate that a requested entry is not found in storage.
type EntryNotFound struct {
	Slug string
}

// IsEntryNotFound returns true if err is an EntryNotFound error.
func IsEntryNotFound(err error) bool {
	if err != nil {
		if _, notFound := err.(EntryNotFound); notFound {
			return true
		}
	}
	return false
}

// Error implements the error interface.
func (e EntryNotFound) Error() string {
	return fmt.Sprintf("entry %s not found", e.Slug)
}
