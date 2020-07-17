/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package app

import (
	"fmt"
	"io/ioutil"
	"log"
	"memory/app/config"
	"memory/app/model"
	"os"
	"testing"
)

func clearData() {
	data.Names = make(map[string]model.Entry)
}

func generateData() {
	clearData()
	for i := 1; i <= 50; i++ {
		tags := []string{"All"}
		if i%2 == 0 {
			tags = append(tags, "even")
		} else {
			tags = append(tags, "odd")
		}
		if i%3 == 0 {
			tags = append(tags, "bythree")
		}
		name := fmt.Sprintf("note #%d", i)
		desc := fmt.Sprintf("note desc #%d", i)
		note := model.NewNote(name, desc, tags)
		data.Names[note.Name()] = note
	}
}

func setupCrud() {
	clearData()
	for i := 0; i < 10; i++ {
		num := i + 1
		note := model.NewNote(fmt.Sprintf("note #%d", num), fmt.Sprintf("desc #%d", num), []string{})
		data.Names[note.Name()] = note
	}
}

func TestGetEntries(t *testing.T) {
	generateData()
	// defaults
	results := GetEntries(EntryTypes{Note: true}, "", "", "", []string{}, 0, 0)
	if len(results.Entries) != 50 {
		t.Errorf("Expected 50 entries, got %d", len(results.Entries))
		return
	}
	if results.Entries[9].Name() != "note #41" {
		t.Errorf("Expected 'note #41', got '%s'", results.Entries[9].Name())
		return
	}
	// no types selected
	results = GetEntries(EntryTypes{}, "", "", "", []string{}, 0, 0)
	if len(results.Entries) != 50 {
		t.Errorf("Expected 50 entries, got %d", len(results.Entries))
		return
	}
	// filter by 1 tag and sort by name
	results = GetEntries(EntryTypes{Note: true}, "", "", "", []string{"odd"}, SortName, 50)
	if len(results.Entries) != 25 {
		t.Errorf("Expected 25 entries, got %d", len(results.Entries))
		return
	}
	if results.Entries[1].Name() != "note #11" {
		t.Errorf("Expected 'note #11', got '%s'", results.Entries[1].Name())
		return
	}
	// filter by 2 tags, sort recent, limit 5
	results = GetEntries(EntryTypes{Note: true}, "", "", "", []string{"odd", "bythree"}, SortRecent, 5)
	if len(results.Entries) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(results.Entries))
		return
	}
	if results.Entries[1].Name() != "note #48" {
		t.Errorf("Expected 'note #48', got '%s'", results.Entries[1].Name())
		return
	}
}

func TestGetEntry(t *testing.T) {
	generateData()
	entry, exists := GetEntry("note #42")
	if !exists {
		t.Error("Unexpected entry not found")
	}
	if entry.Name() != "note #42" {
		t.Error("Expected 'note #42', got", entry.Name())
	}
	entry, exists = GetEntry("invalid")
	if exists {
		t.Error("Expected nil entry, got", entry.Name())
	}
}

// GetNote retrieves and returns the specified note from the collection.
func TestGetNote(t *testing.T) {
	setupCrud()
	var entry model.Entry
	var note model.Note
	var exists bool
	entry, exists = GetEntry("note #3")
	note = entry.(model.Note)
	if !exists {
		t.Error("Unexpected not exists")
	} else if note.Name() != "note #3" || note.Description() != "desc #3" {
		t.Error("Did not get expected note name (test #3) or description (desc #3):", note.Name(), ",", note.Description())
	}
	_, exists = GetEntry("not found")
	if exists {
		t.Error("Expected exists for invalid note name")
	}
}

// PutNote adds or replaces the given note in the collection.
func TestPutNote(t *testing.T) {
	setupCrud()
	newNote := model.NewNote("new note", "", []string{})
	PutEntry(newNote)
	if len(data.Names) != 11 {
		t.Errorf("Expected 11 notes (1st pass), found %d", len(data.Names))
	}
	existingNote := model.NewNote("note #3", "different desc", []string{})
	PutEntry(existingNote)
	if len(data.Names) != 11 {
		t.Errorf("Expected 11 notes (2nd pass), found %d", len(data.Names))
	}
	gotNote, exists := GetEntry("note #3")
	if !exists {
		t.Error("updated note does not exist")
	} else if gotNote.Description() != "different desc" {
		t.Error("Expected 'different desc', got", gotNote.Description())
	}
}

// DeleteNote removes the specified note from the collection.
func TestDeleteNote(t *testing.T) {
	setupCrud()
	existed := DeleteEntry("note #3")
	if !existed {
		t.Error("Note did not exist")
	}
	if len(data.Names) != 9 {
		t.Errorf("Expected 9 notes, got %d", len(data.Names))
	}
	gotNote, exists := GetEntry("note #3")
	if exists {
		t.Error("Deleted note exists")
	}
	if gotNote != nil {
		t.Error("Expected nil, got", gotNote.Name())
	}
}

func TestSave(t *testing.T) {
	setupCrud()
	file, err := ioutil.TempFile(".", "prefix")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	config.MemoryHome = "."
	config.DataFile = file.Name()
	Save()
	data.Names = make(map[string]model.Entry)
	Init()
	if len(data.Names) != 10 {
		t.Error("Expected 10 entries, got", len(data.Names))
	}
}

func TestRename(t *testing.T) {
	setupCrud()
	newName := "renamed note #3"
	err := RenameEntry("note #3", newName)
	if err != nil {
		t.Error(err)
		return
	}
	entry, exists := GetEntry(newName)
	if !exists {
		t.Error("Renamed note doesn't exist")
		return
	} else if entry.Name() != newName {
		t.Errorf("Expected '%s', got '%s", newName, entry.Name())
		return
	}
	if len(data.Names) != 10 {
		t.Error("Expected 10 entries, got", len(data.Names))
	}
}

func TestEdit(t *testing.T) {
	setupCrud()
	entry, exists := GetEntry("note #3")
	if !exists {
		t.Error("note #3 doesn't exist, but should")
	}
	switch typedEntry := entry.(type) {
	case model.Note:
		typedEntry.SetDescription("different")
		PutEntry(typedEntry)
	default:
		t.Error("Unexpected entry type")
	}
	entry2, exists := GetEntry("note #3")
	if !exists {
		t.Error("note #3 doesn't exist (2nd), but should")
	}
	if entry2.Description() != "different" {
		t.Errorf("Expected '%s', got '%s'", "different", entry2.Description())
	}
}
