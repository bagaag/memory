/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package app

import (
	"io/ioutil"
	"memory/app/model"
	"memory/util"
	"testing"
)

func TestParseLinks(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test_parse_links")
	if err != nil {
		t.Error(err)
		return
	}
	Init(tempDir)
	defer util.DelTree(tempDir)
	n1 := model.NewEntry(model.EntryTypeNote, "Exists", "", []string{})
	PutEntry(n1)
	n2 := model.NewEntry(model.EntryTypeNote, "Exists 2", "", []string{})
	PutEntry(n2)
	Save()
	testParseLinks(t, 1, "[Exists]", "[Exists]", []string{"exists"})
	testParseLinks(t, 2, "text [Exists]", "text [Exists]", []string{"exists"})
	testParseLinks(t, 3, "[Exists] text", "[Exists] text", []string{"exists"})
	// we record links to pages that don't exist so they can be listed as broken links
	testParseLinks(t, 4, "[Not Exists]", "[?Not Exists]", []string{"not-exists"})
	testParseLinks(t, 5, "[Exists] [Exists  \n2]", "[Exists] [Exists  \n2]", []string{"exists", "exists-2"})
	testParseLinks(t, 6, "", "", []string{})
	testParseLinks(t, 7, "[Exists]\n[Exists 2]", "[Exists]\n[Exists 2]", []string{"exists", "exists-2"})
	testParseLinks(t, 8, "[~Exists]", "[~Exists]", []string{})
	testParseLinks(t, 9, "[?Exists]", "[Exists]", []string{"exists"})
	testParseLinks(t, 10, "[?Not Exists]", "[?Not Exists]", []string{"not-exists"})
	testParseLinks(t, 11, "[~Not Exists]", "[~Not Exists]", []string{})
	testParseLinks(t, 12, "[Exists 2]\n[Exists]", "[Exists 2]\n[Exists]", []string{"exists-2", "exists"})
	testParseLinks(t, 13, "[Exists](external)", "[Exists](external)", []string{})
}

func testParseLinks(t *testing.T, testNo int, input string, parsedExpected string, linksExpected []string) {
	parsed, links := ParseLinks(input)
	if parsed != parsedExpected {
		t.Errorf("#%d Expected parsed '%s', got '%s'", testNo, parsedExpected, parsed)
	}
	if !util.StringSlicesEqual(linksExpected, links) {
		t.Errorf("#%d Expected links '%s', got '%s'", testNo, linksExpected, links)
	}
}

func TestResolveLinks(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test_resolve_links")
	if err != nil {
		t.Error(err)
		return
	}
	Init(tempDir)
	defer util.DelTree(tempDir)
	n1 := model.NewEntry(model.EntryTypeNote, "Exists", "", []string{})
	PutEntry(n1)
	n2 := model.NewEntry(model.EntryTypeNote, "Exists 2", "", []string{})
	PutEntry(n2)
	Save()
	links := []string{"exists", "not-exists", "exists-2"}
	resolved := ResolveLinks(links)
	if len(resolved) != 2 {
		t.Error("Expected len of 2, got", len(resolved))
	}
	if resolved[1].Name != "Exists 2" {
		t.Error("Expected 'Exists 2', got", resolved[1].Name)
	}
}

func TestPopulateLinks(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test_populate_links")
	if err != nil {
		t.Error(err)
		return
	}
	Init(tempDir)
	defer util.DelTree(tempDir)
	nA := model.NewEntry(model.EntryTypeNote, "Note 1", "This note has a link to [Note 2].", []string{})
	nB := model.NewEntry(model.EntryTypeNote, "Note 2", "This note has a link to [Note 3] and [Note 2].", []string{})
	nC := model.NewEntry(model.EntryTypeNote, "Note 3", "This note has no links.", []string{})
	PutEntry(nA)
	PutEntry(nB)
	PutEntry(nC)
	populateLinks()
	n1, _ := GetEntryFromIndex(util.GetSlug("Note 1"))
	n2, _ := GetEntryFromIndex(util.GetSlug("Note 2"))
	n3, _ := GetEntryFromIndex(util.GetSlug("Note 3"))
	// test linksTo
	if !util.StringSlicesEqual(n1.LinksTo, []string{"note-2"}) {
		t.Error("Expected n1.LinksTo==['note-2'], got", n1.LinksTo)
	}
	if !util.StringSlicesEqual(n2.LinksTo, []string{"note-3", "note-2"}) {
		t.Error("Expected n2.LinksTo==['note-3','note-2'], got", n2.LinksTo)
	}
	if !util.StringSlicesEqual(n3.LinksTo, []string{}) {
		t.Error("Expected n3.LinksTo==[], got", n3.LinksTo)
	}
	// test linkedFrom
	if !util.StringSlicesEqual(n1.LinkedFrom, []string{}) {
		t.Error("Expected n1.LinkedFrom==[], got", n1.LinkedFrom)
	}
	if !util.StringSlicesEqual(n2.LinkedFrom, []string{"note-1", "note-2"}) {
		t.Error("Expected n2.LinkedFrom==['note-1','note-2'], got", n2.LinkedFrom)
	}
	if !util.StringSlicesEqual(n3.LinkedFrom, []string{"note-2"}) {
		t.Error("Expected n3.LinkedFrom==['note-2'], got", n3.LinkedFrom)
	}
}

func TestBrokenLinks(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test_broken_links")
	if err != nil {
		t.Error(err)
		return
	}
	Init(tempDir)
	defer util.DelTree(tempDir)
	nA := model.NewEntry(model.EntryTypeNote, "Note 1", "This note has a link to [Note A].", []string{})
	nB := model.NewEntry(model.EntryTypeNote, "Note 2", "This note [has a] link to [note 4] and [Note 1].", []string{})
	nC := model.NewEntry(model.EntryTypeNote, "Note 3", "This note has no links.", []string{})
	PutEntry(nA)
	PutEntry(nB)
	PutEntry(nC)
	populateLinks()
	entriesWithBL, err := BrokenLinks()
	if err != nil {
		t.Error(err)
	}
	if !util.StringSlicesEqual(entriesWithBL["Note 1"], []string{"note-a"}) {
		t.Errorf("Expected %s, got %s", []string{"note-a"}, entriesWithBL["Note 1"])
	}
	if !util.StringSlicesEqual(entriesWithBL["Note 2"], []string{"has-a", "note-4"}) {
		t.Errorf("Expected %s, got %s", []string{"has-a", "note-4"}, entriesWithBL["Note 2"])
	}
	if !util.StringSlicesEqual(entriesWithBL["Note 3"], []string{}) {
		t.Errorf("Expected %s, got %s", []string{}, entriesWithBL["Note 1"])
	}
}
