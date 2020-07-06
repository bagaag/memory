/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package model

import (
  "testing"
)

func TestNewNote(t *testing.T) {
  desc := "descr"
  tags := []string{"tag1","tag2"}
  n := NewNote(desc, tags)
  if n.Description != desc || len(n.Tags) != 2 {
    t.Errorf("new note attributes not set: %s", n)
  }
  if len(notes) != 1 {
    t.Errorf("expect len(notes)==1, got %d", len(notes))
  }
}
