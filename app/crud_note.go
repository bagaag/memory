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
	"memory/app/model"
)

// GetNote retrieves and returns the specified note from the collection.
func GetNote(id string) model.Note {
	//TODO: implement GetNote
	return model.NewNote("not implemented", "", []string{})
}

// PutNote adds or replaces the given note in the collection.
func PutNote(note model.Note) {
	//TODO: check for existing name in PutNote
	data.Notes = append(data.Notes, note)
}

// DeleteNote removes the specified note from the collection.
func DeleteNote(id string) {
}
