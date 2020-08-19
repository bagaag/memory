/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package app

import (
	"memory/app/config"
	"memory/app/persist"
	"memory/util"
	"testing"
)

/* This file contains functions to support full text entry search. */

func TestSearch(t *testing.T) {
	e1 := NewEntry(EntryTypeNote, "Apple Heresay", "Yours is no disgrace.", []string{"tag1"})
	e2 := NewEntry(EntryTypeNote, "Bungled Apple", "Shaky groove turtle.", []string{"tag2", "tag1"})
	e3 := NewEntry(EntryTypeNote, "Frenetic Plum", "Undersea groove turntable swing.", []string{"tag3"})
	data.Names[e2.Slug()] = e2
	data.Names[e1.Slug()] = e1
	data.Names[e3.Slug()] = e3
	config.MemoryHome = "/tmp"
	if err := initSearch(); err != nil {
		t.Error(err)
	}
	// name search
	r, err := executeSearch("apple")
	if err != nil {
		t.Error(err)
	}
	expect := []string{"apple-heresay", "bungled-apple"}
	if !util.StringSlicesEqual(r, expect) {
		t.Errorf("1. Expected %v, got %v", expect, r)
	}
	// tag search
	r, err = executeSearch("tag1")
	if err != nil {
		t.Error(err)
	}
	expect = []string{"apple-heresay", "bungled-apple"}
	if !util.StringSlicesEqual(r, expect) {
		t.Errorf("2. Expected %v, got %v", expect, r)
	}
	// description search
	r, err = executeSearch("groove turtle")
	if err != nil {
		t.Error(err)
	}
	expect = []string{"bungled-apple"}
	if !util.StringSlicesEqual(r, expect) {
		t.Errorf("3. Expected %v, got %v", expect, r)
	}
	// cleanup
	persist.RemoveFile(config.SearchPath())
}
