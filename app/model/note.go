/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package model

import (
	"encoding/json"
	"time"
)

// Note provides a temporary holding place where ideas that
// require further development can be quickly captured.
type Note struct {
	Entry
}

// NoteJSON provides public fields for JSON marshalling.
type NoteJSON struct {
	Name        string
	Description string
	Tags        []string
	LinksTo     []string
	LinkedFrom  []string
	Created     time.Time
	Modified    time.Time
}

// NewNote creates and returns a Note object
func NewNote(name string, description string, tags []string) Note {
	now := time.Now()
	note := Note{
		Entry: Entry{
			name:        name,
			description: description,
			tags:        tags,
			linksTo:     []string{},
			linkedFrom:  []string{},
			created:     now,
			modified:    now,
		},
	}
	return note
}

// Name getter
func (note Note) Name() string {
	return note.name
}

// SetName setter
func (note *Note) SetName(name string) {
	note.name = name
}

// Description getter
func (note Note) Description() string {
	return note.description
}

// SetDescription setter
func (note *Note) SetDescription(desc string) {
	note.description = desc
}

// Tags getter
func (note Note) Tags() []string {
	return note.tags
}

// SetTags setter
func (note *Note) SetTags(tags []string) {
	note.tags = tags
}

// LinksTo contains entry names this entry's description contains links to.
func (note Note) LinksTo() []string {
	return note.linksTo
}

// SetLinksTo setter
func (note *Note) SetLinksTo(names []string) {
	note.linksTo = names
}

// LinkedFrom contains entry names that link to this entry.
func (note Note) LinkedFrom() []string {
	return note.linkedFrom
}

// SetLinkedFrom setter
func (note *Note) SetLinkedFrom(names []string) {
	note.linkedFrom = names
}

// Created getter
func (note Note) Created() time.Time {
	return note.created
}

// Modified getter
func (note Note) Modified() time.Time {
	return note.modified
}

// MarshalJSON translates private fields to public fields for JSON storage and retrieval.
func (note Note) MarshalJSON() ([]byte, error) {
	j, err := json.Marshal(NoteJSON{
		Name:        note.name,
		Description: note.description,
		Tags:        note.tags,
		Created:     note.created,
		Modified:    note.modified,
	})
	if err != nil {
		return nil, err
	}
	return j, nil
}

// UnmarshalJSON translates public fields to public fields for JSON storage and retrieval.
func (note *Note) UnmarshalJSON(b []byte) error {
	noteJSON := NoteJSON{}
	if err := json.Unmarshal(b, &noteJSON); err != nil {
		return err
	}
	note.name = noteJSON.Name
	note.description = noteJSON.Description
	note.tags = noteJSON.Tags
	note.created = noteJSON.Created
	note.modified = noteJSON.Modified
	return nil
}
