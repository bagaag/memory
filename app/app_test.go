/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package app

import (
	"fmt"
	"memory/app/model"
	"testing"
)

func clearData() {
	data.Notes = []model.Note{}
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
		data.Notes = append(data.Notes, note)
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
	if len(results.Entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(results.Entries))
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
	entry, err := GetEntry("noTe", "note #42")
	if err != nil {
		t.Error("Unexpected error getting entry:", err)
	}
	if entry.Name() != "note #42" {
		t.Error("Expected 'note #42', got", entry.Name())
	}
	entry, err = GetEntry("invalid", "invalid")
	if entry != nil {
		t.Error("Expected nil entry, got", entry.Name())
	}
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
