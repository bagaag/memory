/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package app

import (
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
	Type        EntryType
	Start       string // Events
	End         string // Events
	Latitude    string // Place
	Longitude   string // Place
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
		Created:     now,
		Modified:    now,
		Type:        entryType,
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

// EntryTypeNotePlural indicates multiple notes.
const EntryTypeNotePlural = "Notes"

// EntryTypeEvent indicates an event.
const EntryTypeEvent = "Event"

// EntryTypeEventPlural indicates multiple events.
const EntryTypeEventPlural = "Events"

// EntryTypePerson indicates a person.
const EntryTypePerson = "Person"

// EntryTypePersonPlural indicates multiple persons
const EntryTypePersonPlural = "People"

// EntryTypePlace indicates a place.
const EntryTypePlace = "Place"

// EntryTypePlacePlural indicates multiple places.
const EntryTypePlacePlural = "Places"

// EntryTypeThing indicates a thing.
const EntryTypeThing = "Thing"

// EntryTypeThingPlural indicates multiple things.
const EntryTypeThingPlural = "Things"

// TagsString returns the entrys tags as a comma-separated string.
func (entry Entry) TagsString() string {
	return strings.Join(entry.Tags, ",")
}
