/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/*
CRUD methods for Notes
*/

package app

import (
	"errors"
	"memory/app/model"
)

// GetNote retrieves and returns the specified note from the collection.
func GetNote(name string) (model.Note, error) {
	for _, note := range data.Notes {
		if note.Name() == name {
			return note, nil
		}
	}
	return model.Note{}, errors.New("There is no note named '" + name + "'.")
}

// PutNote adds or replaces the given note in the collection.
func PutNote(note model.Note) {
	for ix, n := range data.Notes {
		if n.Name() == note.Name() {
			data.Notes[ix] = note
			return
		}
	}
	data.Notes = append(data.Notes, note)
}

// DeleteNote removes the specified note from the collection.
func DeleteNote(name string) error {
	for ix, note := range data.Notes {
		if note.Name() == name {
			data.Notes = append(data.Notes[:ix], data.Notes[ix+1:]...)
			// TODO: if we're saving data in a session, need to update or clear for entry deletion
			return nil
		}
	}
	return errors.New("There is no note named '" + name + "'.")
}
