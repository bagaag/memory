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
	data.Notes = data.Notes[0:0]
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
		PutNote(note)
	}
}

func TestGetEntries(t *testing.T) {
	generateData()
	// defaults
	notes := GetEntries(EntryTypes{Note: true}, "", "", "", []string{}, 0, 0)
	if len(notes) != 50 {
		t.Errorf("Expected 50 entries, got %d", len(notes))
		return
	}
	if notes[9].Name() != "note #41" {
		t.Errorf("Expected 'note #41', got '%s'", notes[9].Name())
		return
	}
	// no types selected
	notes = GetEntries(EntryTypes{}, "", "", "", []string{}, 0, 0)
	if len(notes) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(notes))
		return
	}
	// filter by 1 tag and sort by name
	notes = GetEntries(EntryTypes{Note: true}, "", "", "", []string{"odd"}, SortName, 50)
	if len(notes) != 25 {
		t.Errorf("Expected 25 entries, got %d", len(notes))
		return
	}
	if notes[1].Name() != "note #11" {
		t.Errorf("Expected 'note #11', got '%s'", notes[1].Name())
		return
	}
	// filter by 2 tags, sort recent, limit 5
	notes = GetEntries(EntryTypes{Note: true}, "", "", "", []string{"odd", "bythree"}, SortRecent, 5)
	if len(notes) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(notes))
		return
	}
	if notes[1].Name() != "note #48" {
		t.Errorf("Expected 'note #48', got '%s'", notes[1].Name())
		return
	}
}
