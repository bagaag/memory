/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

// The persist package handles persistence tasks.

package persist

import (
	"memory/app/model"
)

//TODO: Move simple persist impl outside app pkg and move Persister to app pkg
// Persist is an interface used to support pluggable entry persistence implementations.
type Persister interface {
	// ReadEntry returns an Entry identified by slug populated from storage.
	ReadEntry(slug string) (model.Entry, error)
	// EntrySlugs returns a string slice containing the slug of every entry in storage.
	EntrySlugs() ([]string, error)
	// EntryExists returns true if the entry is found in storage
	EntryExists(slug string) bool
	// SaveEntry writes the entry to storage.
	SaveEntry(entry model.Entry) error
	// DeleteEntry removes the entry idenfied by slug from storage.
	DeleteEntry(slug string) error
	// RenameEntry moves an entry from one slug to another, reflecting a new name
	RenameEntry(oldName string, newName string) (model.Entry, error)
}
