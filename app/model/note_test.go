/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package model

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestNewNote(t *testing.T) {
	name := "a note"
	desc := "the note text"
	tags := []string{"tag1", "tag2"}
	n := NewNote(name, desc, tags)
	if n.description != desc || len(n.tags) != 2 {
		t.Errorf("new note attributes not set")
	}
}

func TestJsonMarshal(t *testing.T) {
	note := NewNote("a note", "the note text", []string{"tag1", "tag2"})
	if ba, err := json.Marshal(note); err != nil {
		t.Errorf("marshal failed: %s", err)
	} else {
		fmt.Println("Json bytes:", string(ba))
		note2 := Note{}
		if err := json.Unmarshal(ba, &note2); err != nil {
			t.Errorf("marshal failed: %s", err)
		} else {
			if note.name != note2.name || note.created.Format("1/2/2006") != note2.created.Format("1/2/2006") {
				t.Errorf("unmarshalled fields don't match [%s != %s || %s != %s]",
					note.name, note2.name, note.created.Format("1/2/2006"), note2.created.Format("1/2/2006"))
			}
		}
	}
}
