/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/*
Test CRUD methods for Notes
*/

package app

import (
	"fmt"
	"memory/app/model"
	"testing"
)

func setup() {
	data.Notes = []model.Note{}
	for i := 0; i < 10; i++ {
		num := i + 1
		note := model.NewNote(fmt.Sprintf("note #%d", num), fmt.Sprintf("desc #%d", num), []string{})
		data.Notes = append(data.Notes, note)
	}
}

// GetNote retrieves and returns the specified note from the collection.
func TestGetNote(t *testing.T) {
	setup()
	var note model.Note
	var err error
	note, err = GetNote("note #3")
	if err != nil {
		t.Error(err)
	} else if note.Name() != "note #3" || note.Description() != "desc #3" {
		t.Error("Did not get expected note name (test #3) or description (desc #3):", note.Name(), ",", note.Description())
	}
	note, err = GetNote("not found")
	if err == nil {
		t.Error("Expected error for invalid note name")
	}
}

// PutNote adds or replaces the given note in the collection.
func TestPutNote(t *testing.T) {
	setup()
	newNote := model.NewNote("new note", "", []string{})
	PutNote(newNote)
	if len(data.Notes) != 11 {
		t.Errorf("Expected 11 notes (1st pass), found %d", len(data.Notes))
	}
	existingNote := model.NewNote("note #3", "different desc", []string{})
	PutNote(existingNote)
	if len(data.Notes) != 11 {
		t.Errorf("Expected 11 notes (2nd pass), found %d", len(data.Notes))
	}
	gotNote, err := GetNote("note #3")
	if err != nil {
		t.Error("Error getting updated note:", err)
	} else if gotNote.Description() != "different desc" {
		t.Error("Expected 'different desc', got", gotNote.Description())
	}
}

// DeleteNote removes the specified note from the collection.
func TestDeleteNote(t *testing.T) {
	setup()
	err := DeleteNote("note #3")
	if err != nil {
		t.Error("Error deleting note:", err)
	}
	if len(data.Notes) != 9 {
		t.Errorf("Expected 9 notes, got %d", len(data.Notes))
	}
	gotNote, err := GetNote("note #3")
	if err == nil {
		t.Error("Expected error getting deleted note")
	}
	if gotNote.Name() != "" {
		t.Error("Expected empty name, got", gotNote.Name())
	}
}
