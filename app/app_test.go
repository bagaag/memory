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
	"memory/util"
	"os"
	"testing"
)

var tempDir1 string
var tempDir2 string

func setupTeardown1(t *testing.T, teardown bool) {
	if !teardown {
		// setup
		var err error
		tempDir1, err = ioutil.TempDir("", "app_test")
		if err != nil {
			t.Error(err)
			return
		}
		Init(tempDir1)
		for i := 1; i <= 50; i++ {
			tags := []string{"all"}
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
			note := NewEntry(EntryTypeNote, name, desc, tags)
			PutEntry(note)
		}
		Save()
	} else {
		// teardown
		err := util.DelTree(tempDir1)
		if err != nil {
			t.Error(err)
		}
	}
}

func setupTeardown2(t *testing.T, teardown bool) {
	if !teardown {
		// setup
		var err error
		tempDir2, err = ioutil.TempDir("", "app_test")
		if err != nil {
			t.Error(err)
			return
		}
		Init(tempDir2)
		for i := 0; i < 10; i++ {
			num := i + 1
			note := NewEntry(EntryTypeNote, fmt.Sprintf("note #%d", num), fmt.Sprintf("desc #%d", num), []string{})
			PutEntry(note)
		}
		Save()
	} else {
		// teardown
		err := util.DelTree(tempDir2)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGetEntry(t *testing.T) {
	setupTeardown1(t, false)
	defer setupTeardown1(t, true)
	entry, exists := GetEntryFromIndex(GetSlug("note #42"))
	if !exists {
		t.Error("Unexpected entry not found")
	}
	if entry.Name != "note #42" {
		t.Error("Expected 'note #42', got", entry.Name)
	}
	entry, exists = GetEntryFromIndex("invalid")
	if exists {
		t.Error("Expected nil entry, got", entry.Name)
	}
}

// GetNote retrieves and returns the specified note from the collection.
func TestGetNote(t *testing.T) {
	setupTeardown2(t, false)
	defer setupTeardown2(t, true)
	var entry Entry
	var note Entry
	var exists bool
	entry, exists = GetEntryFromIndex(GetSlug("note #3"))
	note = entry
	if !exists {
		t.Error("Unexpected not exists")
	} else if note.Name != "note #3" || note.Description != "desc #3" {
		t.Error("Did not get expected note name (test #3) or description (desc #3):", note.Name, ",", note.Description)
	}
	_, exists = GetEntryFromIndex("not found")
	if exists {
		t.Error("Expected exists for invalid note name")
	}
}

// PutNote adds or replaces the given note in the collection.
func TestPutNote(t *testing.T) {
	setupTeardown2(t, false)
	defer setupTeardown2(t, true)
	newNote := NewEntry(EntryTypeNote, "new note", "", []string{})
	PutEntry(newNote)
	if len(data.Names) != 1 {
		t.Errorf("Expected 1 notes (1st pass), found %d", len(data.Names))
	}
	existingNote := NewEntry(EntryTypeNote, "note #3", "different desc", []string{})
	PutEntry(existingNote)
	if len(data.Names) != 2 {
		t.Errorf("Expected 2 notes (2nd pass), found %d", len(data.Names))
	}
	Save()
	gotNote, exists := GetEntryFromIndex(GetSlug("note #3"))
	if !exists {
		t.Error("updated note does not exist")
	} else if gotNote.Description != "different desc" {
		t.Error("Expected 'different desc', got", gotNote.Description)
	}
}

// DeleteNote removes the specified note from the collection.
func TestDeleteNote(t *testing.T) {
	setupTeardown2(t, false)
	defer setupTeardown2(t, true)
	existed := DeleteEntry(GetSlug("note #3"))
	if !existed {
		t.Error("Note did not exist")
	}
	if IndexedCount() != 9 {
		t.Errorf("Expected 9 notes, got %d", len(data.Names))
	}
	_, exists := GetEntryFromIndex(GetSlug("note #3"))
	if exists {
		t.Error("Deleted note exists")
	}
}

func TestSave(t *testing.T) {
	setupTeardown2(t, false)
	defer setupTeardown2(t, true)
	file, err := ioutil.TempFile(".", "TestSave")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	config.MemoryHome = "."
	config.DataFile = file.Name()
	Save()
	data.Names = make(map[string]Entry)
	Init("")
	if len(data.Names) != 10 {
		t.Error("Expected 10 entries, got", len(data.Names))
	}
	entry, exists := GetEntryFromIndex("note #3")
	if !exists {
		t.Error("Expected note #3 to exist, it does not")
	}
	if entry.Name != "note #3" {
		t.Error("Expected 'note #3', got", entry.Name)
	}
}

func TestRename(t *testing.T) {
	setupTeardown2(t, false)
	defer setupTeardown2(t, true)
	newName := "renamed note #3"
	err := RenameEntry("note #3", newName)
	if err != nil {
		t.Error(err)
		return
	}
	entry, exists := GetEntryFromIndex(newName)
	if !exists {
		t.Error("Renamed note doesn't exist")
		return
	} else if entry.Name != newName {
		t.Errorf("Expected '%s', got '%s", newName, entry.Name)
		return
	}
	if len(data.Names) != 10 {
		t.Error("Expected 10 entries, got", len(data.Names))
	}
}

func TestEdit(t *testing.T) {
	setupTeardown2(t, false)
	defer setupTeardown2(t, true)
	entry, exists := GetEntryFromIndex("note #3")
	if !exists {
		t.Error("note #3 doesn't exist, but should")
	}
	entry.Description = "different"
	PutEntry(entry)
	entry2, exists := GetEntryFromIndex("note #3")
	if !exists {
		t.Error("note #3 doesn't exist (2nd), but should")
	}
	if entry2.Description != "different" {
		t.Errorf("Expected '%s', got '%s'", "different", entry2.Description)
	}
}
