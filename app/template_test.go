/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/* Tests for the template functions. */

package app

import (
	"memory/util"
	"testing"
)

func TestParseYamlDown(t *testing.T) {
	s := `---
Type: Note
Name: Note #1
Tags: one,two,three
---

Hey now. This is the description.
`
	entry, err := ParseYamlDown(s)
	if err != nil {
		t.Error(err)
	} else {
		if entry.Name != "Note #1" {
			t.Error("Expected 'Note #1', got", entry.Name)
		}
		if entry.Type != EntryTypeNote {
			t.Error("Expected 'Note', got", entry.Type)
		}
		if !util.StringSlicesEqual(entry.Tags, []string{"one", "two", "three"}) {
			t.Error("Expected 'one,two,three', got", entry.Tags)
		}
		if entry.Description != "Hey now. This is the description." {
			t.Error("Expected 'Hey now. ...', got ", entry.Description)
		}
	}
}

func TestParseYamlDownPlace(t *testing.T) {
	s := `---
Type: Place
Name: Place #1
Tags: one,two,three
Latitude: 42.468230
Longitude: -71.436690
Address: 24 Oakwood Rd, Acton, MA
---

Hey now. This is the description.
`
	entry, err := ParseYamlDown(s)
	if err != nil {
		t.Error(err)
	} else {
		if entry.Latitude != "42.468230" {
			t.Error("Expected '42.468230', got", entry.Latitude)
		}
		if entry.Longitude != "-71.436690" {
			t.Error("Expected '-71.436690', got", entry.Longitude)
		}
		if entry.Address != "24 Oakwood Rd, Acton, MA" {
			t.Error("Expected '24 Oakwood Rd, Acton, MA', got", entry.Address)
		}
	}
}

func TestRenderYamlDown(t *testing.T) {
	entry := NewEntry(EntryTypeNote, "Note #1", "Hey now. This is the description.", []string{"one", "two", "three"})
	entry.Custom["Custom 1"] = "Custom Value 1"
	expect := `---
Name: Note #1
Type: Note
Tags: one,two,three
Custom 1: Custom Value 1
---

Hey now. This is the description.
`
	s, err := RenderYamlDown(entry)
	if err != nil {
		t.Error(err)
	}
	if s != expect {
		t.Error("Unexpected result:", s)
	}
}

func TestRenderYamlDownPlace(t *testing.T) {
	entry := Entry{
		Type:        EntryTypePlace,
		Name:        "Place #1",
		Description: "Hey now.",
		Address:     "Addr 1",
		Latitude:    "42.468230",
		Longitude:   "-71.436690",
		Custom:      make(map[string]string),
	}
	expect := `---
Name: Place #1
Type: Place
Tags: 
Address: Addr 1
Latitude: 42.468230
Longitude: -71.436690
---

Hey now.
`
	s, err := RenderYamlDown(entry)
	if err != nil {
		t.Error(err)
	}
	if s != expect {
		t.Error("Unexpected result:", s)
	}
}
