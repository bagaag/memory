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

func TestRenderYamlDown(t *testing.T) {
	entry := NewEntry(EntryTypeNote, "Note #1", "Hey now. This is the description.", []string{"one", "two", "three"})
	expect := `---
Name: Note #1
Type: Note
Tags: one,two,three
---
`

Hey now. This is the description.
	s, err := RenderYamlDown(entry)
	if err != nil {
		t.Error(err)
	}
	if s != expect {
		t.Error("Unexpected result:", s)
	}
}
