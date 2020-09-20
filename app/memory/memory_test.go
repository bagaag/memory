/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package memory

import (
	"fmt"
	"io/ioutil"
	"memory/app/model"
	"memory/util"
	"testing"
)

var tempDir1 string
var tempDir2 string

func setupTeardown1(t *testing.T, teardown bool) *Memory {
	var memApp *Memory
	if !teardown {
		// setup
		var err error
		tempDir1, err = ioutil.TempDir("", "app_test")
		if err != nil {
			t.Error(err)
			return memApp
		}
		memApp, err = Init(tempDir1)
		if err != nil {
			t.Error(err)
			return memApp
		}
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
			note := model.NewEntry(model.EntryTypeNote, name, desc, tags)
			memApp.PutEntry(note)
		}
	} else {
		// teardown
		err := util.DelTree(tempDir1)
		if err != nil {
			t.Error(err)
		}
	}
	return memApp
}

func setupTeardown2(t *testing.T, teardown bool) *Memory {
	var memApp *Memory
	if !teardown {
		// setup
		var err error
		tempDir2, err = ioutil.TempDir("", "app_test")
		if err != nil {
			t.Error(err)
			return memApp
		}
		memApp, err = Init(tempDir2)
		if err != nil {
			t.Error(err)
			return memApp
		}
		for i := 0; i < 10; i++ {
			num := i + 1
			note := model.NewEntry(model.EntryTypeNote, fmt.Sprintf("note #%d", num), fmt.Sprintf("desc #%d", num), []string{})
			memApp.PutEntry(note)
		}
	} else {
		// teardown
		err := util.DelTree(tempDir2)
		if err != nil {
			t.Error(err)
		}
	}
	return memApp
}

func TestGetEntry(t *testing.T) {
	memApp := setupTeardown1(t, false)
	defer setupTeardown1(t, true)
	entry, err := memApp.GetEntry(util.GetSlug("note #42"))
	if err != nil {
		t.Error(err)
	}
	if entry.Name != "note #42" {
		t.Error("Expected 'note #42', got", entry.Name)
	}
	entry, err = memApp.GetEntry("invalid")
	if err == nil || !model.IsEntryNotFound(err) {
		t.Error("Expected nil entry, got", entry.Name, err)
	}
}

// GetNote retrieves and returns the specified note from the collection.
func TestGetNote(t *testing.T) {
	memApp := setupTeardown2(t, false)
	defer setupTeardown2(t, true)
	var entry model.Entry
	var note model.Entry
	var err error
	entry, err = memApp.GetEntry(util.GetSlug("note #3"))
	note = entry
	if err != nil {
		t.Error(err)
	} else if note.Name != "note #3" || note.Description != "desc #3" {
		t.Error("Did not get expected note name (test #3) or description (desc #3):", note.Name, ",", note.Description)
	}
	_, err = memApp.GetEntry("not found")
	if err == nil || !model.IsEntryNotFound(err) {
		t.Error("Expected exists for invalid note name", err)
	}
}

// PutNote adds or replaces the given note in the collection.
func TestPutNote(t *testing.T) {
	memApp := setupTeardown2(t, false)
	defer setupTeardown2(t, true)
	newNote := model.NewEntry(model.EntryTypeNote, "new note", "", []string{})
	memApp.PutEntry(newNote)
	list, err := memApp.Persist.EntrySlugs()
	if err != nil {
		t.Error(err)
		return
	}
	if len(list) != 11 {
		t.Errorf("Expected 11 notes (1st pass), found %d", len(list))
	}
	existingNote := model.NewEntry(model.EntryTypeNote, "note #3", "different desc", []string{})
	memApp.PutEntry(existingNote)
	list, err = memApp.Persist.EntrySlugs()
	if err != nil {
		t.Error(err)
		return
	}
	if len(list) != 11 {
		t.Errorf("Expected 11 notes (2nd pass), found %d", len(list))
	}
	gotNote, err := memApp.GetEntry(util.GetSlug("note #3"))
	if err != nil {
		t.Error("updated note does not exist", err)
	} else if gotNote.Description != "different desc" {
		t.Error("Expected 'different desc', got", gotNote.Description)
	}
}

// DeleteNote removes the specified note from the collection.
func TestDeleteNote(t *testing.T) {
	memApp := setupTeardown2(t, false)
	defer setupTeardown2(t, true)
	err := memApp.DeleteEntry(util.GetSlug("note #3"))
	if err != nil {
		t.Error(err)
	}
	list, err := memApp.Persist.EntrySlugs()
	if err != nil {
		t.Error(err)
		return
	}
	if len(list) != 9 {
		t.Errorf("Expected 9 notes, got %d", len(list))
	}
	_, err = memApp.GetEntry(util.GetSlug("note #3"))
	if err == nil {
		t.Error("Deleted note exists")
	}
}

func TestRename(t *testing.T) {
	memApp := setupTeardown2(t, false)
	defer setupTeardown2(t, true)
	newName := "renamed note #3"
	entry, err := memApp.RenameEntry("note #3", newName)
	if err != nil {
		t.Error(err)
		return
	}
	if err != nil {
		t.Error("Renamed note doesn't exist")
		return
	} else if entry.Name != newName {
		t.Errorf("Expected '%s', got '%s", newName, entry.Name)
		return
	}

	list, err := memApp.Persist.EntrySlugs()
	if err != nil {
		t.Error(err)
		return
	}
	if len(list) != 10 {
		t.Error("Expected 10 entries, got", len(list))
	}
}

func TestEdit(t *testing.T) {
	memApp := setupTeardown2(t, false)
	defer setupTeardown2(t, true)
	entry, err := memApp.GetEntry(util.GetSlug("note #3"))
	if err != nil {
		t.Error("note #3 doesn't exist, but should", err)
	}
	entry.Description = "different"
	memApp.PutEntry(entry)
	entry2, err := memApp.GetEntry(util.GetSlug("note #3"))
	if err != nil {
		t.Error("note #3 doesn't exist (2nd), but should")
	}
	if entry2.Description != "different" {
		t.Errorf("Expected '%s', got '%s'", "different", entry2.Description)
	}
}
