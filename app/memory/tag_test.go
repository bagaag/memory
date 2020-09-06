/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package memory

import (
	"memory/app/model"
	"memory/util"
	"testing"
)

/* This file contains tests for the functions in tag.go. */

func TestGetTags(t *testing.T) {
	memApp := setupTeardown1(t, false)
	defer setupTeardown1(t, true)
	tags, err := memApp.GetTags()
	if err != nil {
		t.Error(err)
	}
	if len(tags) != 4 {
		t.Errorf("Expected 4 tags, got %d", len(tags))
	}
	names := tags["even"]
	if len(names) != 25 {
		t.Errorf("Expected 25 even names, got %d", len(names))
	}
	if names[0] != "note #10" {
		t.Errorf("Expected first even note to be 'note #10', got %s", names[0])
	}
	if names[24] != "note #8" {
		t.Errorf("Expected last even note to be 'note #8', got %s", names[24])
	}
}

func TestGetSortedTags(t *testing.T) {
	memApp := setupTeardown1(t, false)
	defer setupTeardown1(t, true)
	tags, err := memApp.GetTags()
	if err != nil {
		t.Error(err)
	}
	sorted := memApp.GetSortedTags(tags)
	expect := []string{"all", "bythree", "even", "odd"}
	if !util.StringSlicesEqual(sorted, expect) {
		t.Errorf("Expected %s, got %s", expect, sorted)
	}
}

func TestTagMatches(t *testing.T) {
	memApp := setupTeardown2(t, false)
	defer setupTeardown2(t, true)
	entry := model.NewEntry(model.EntryTypeNote, "Test", "Description", []string{"one", "two"})
	if !memApp.tagMatches(entry, []string{"one", "two"}, false) {
		t.Errorf("Failed tagMatches test #1")
	} else if !memApp.tagMatches(entry, []string{"one", "two"}, true) {
		t.Errorf("Failed tagMatches test #2")
	} else if memApp.tagMatches(entry, []string{"three", "four"}, false) {
		t.Errorf("Failed tagMatches test #3")
	} else if memApp.tagMatches(entry, []string{"three", "four"}, true) {
		t.Errorf("Failed tagMatches test #4")
	} else if memApp.tagMatches(entry, []string{"one", "three"}, true) {
		t.Errorf("Failed tagMatches test #5")
	} else if !memApp.tagMatches(entry, []string{"one", "three"}, false) {
		t.Errorf("Failed tagMatches test #6")
	}
}
