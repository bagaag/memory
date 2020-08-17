/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package app

import (
	"memory/util"
	"testing"
)

func TestParseLinks(t *testing.T) {
	n1 := NewEntry(EntryTypeNote, "Exists", "", []string{})
	PutEntry(n1)
	n2 := NewEntry(EntryTypeNote, "Exists 2", "", []string{})
	PutEntry(n2)
	testParseLinks(t, 1, "[Exists]", "[Exists]", []string{"Exists"})
	testParseLinks(t, 2, "text [Exists]", "text [Exists]", []string{"Exists"})
	testParseLinks(t, 3, "[Exists] text", "[Exists] text", []string{"Exists"})
	// we record links to pages that don't exist so they can be listed as broken links
	testParseLinks(t, 4, "[Not Exists]", "[!Not Exists]", []string{"Not Exists"})
	testParseLinks(t, 5, "[Exists] [Exists  \n2]", "[Exists] [Exists  \n2]", []string{"Exists", "Exists 2"})
	testParseLinks(t, 6, "", "", []string{})
	testParseLinks(t, 7, "[Exists]\n[Exists 2]", "[Exists]\n[Exists 2]", []string{"Exists", "Exists 2"})
	testParseLinks(t, 8, "[~Exists]", "[~Exists]", []string{})
	testParseLinks(t, 9, "[!Exists]", "[Exists]", []string{"Exists"})
	testParseLinks(t, 10, "[!Not Exists]", "[!Not Exists]", []string{"Not Exists"})
	testParseLinks(t, 11, "[~Not Exists]", "[~Not Exists]", []string{})
	testParseLinks(t, 12, "[Exists 2]\n[Exists]", "[Exists 2]\n[Exists]", []string{"Exists 2", "Exists"})
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
	n1 := NewEntry(EntryTypeNote, "Exists", "", []string{})
	PutEntry(n1)
	n2 := NewEntry(EntryTypeNote, "Exists 2", "", []string{})
	PutEntry(n2)
	links := []string{"Exists", "Not exists", "Exists 2"}
	resolved := ResolveLinks(links)
	if len(resolved) != 2 {
		t.Error("Expected len of 2, got", len(resolved))
	}
	if resolved[1].Name != "Exists 2" {
		t.Error("Expected 'Exists 2', got", resolved[1].Name)
	}
}

func TestPopulateLinks(t *testing.T) {
	nA := NewEntry(EntryTypeNote, "Note 1", "This note has a link to [Note 2].", []string{})
	nB := NewEntry(EntryTypeNote, "Note 2", "This note has a link to [Note 3] and [Note 2].", []string{})
	nC := NewEntry(EntryTypeNote, "Note 3", "This note has no links.", []string{})
	PutEntry(nA)
	PutEntry(nB)
	PutEntry(nC)
	populateLinks()
	n1, _ := GetEntry(GetSlug("Note 1"))
	n2, _ := GetEntry(GetSlug("Note 2"))
	n3, _ := GetEntry(GetSlug("Note 3"))
	// test linksTo
	if !util.StringSlicesEqual(n1.LinksTo, []string{"Note 2"}) {
		t.Error("Expected n1.LinksTo==['Note 2'], got", n1.LinksTo)
	}
	if !util.StringSlicesEqual(n2.LinksTo, []string{"Note 3", "Note 2"}) {
		t.Error("Expected n2.LinksTo==['Note 3','Note 2'], got", n2.LinksTo)
	}
	if !util.StringSlicesEqual(n3.LinksTo, []string{}) {
		t.Error("Expected n3.LinksTo==[], got", n3.LinksTo)
	}
	// test linkedFrom
	if !util.StringSlicesEqual(n1.LinkedFrom, []string{}) {
		t.Error("Expected n1.LinkedFrom==[], got", n1.LinkedFrom)
	}
	if !util.StringSlicesEqual(n2.LinkedFrom, []string{"Note 1", "Note 2"}) {
		t.Error("Expected n2.LinkedFrom==['Note 1','Note 2'], got", n2.LinkedFrom)
	}
	if !util.StringSlicesEqual(n3.LinkedFrom, []string{"Note 2"}) {
		t.Error("Expected n3.LinkedFrom==['Note 2'], got", n3.LinkedFrom)
	}
}

func TestBrokenLinks(t *testing.T) {
	nA := NewEntry(EntryTypeNote, "Note 1", "This note has a link to [Note A].", []string{})
	nB := NewEntry(EntryTypeNote, "Note 2", "This note [has a] link to [note 4] and [Note 1].", []string{})
	nC := NewEntry(EntryTypeNote, "Note 3", "This note has no links.", []string{})
	PutEntry(nA)
	PutEntry(nB)
	PutEntry(nC)
	populateLinks()
	entriesWithBL := BrokenLinks()
	if !util.StringSlicesEqual(entriesWithBL["Note 1"], []string{"Note A"}) {
		t.Errorf("Expected %s, got %s", []string{"Note A"}, entriesWithBL["Note 1"])
	}
	if !util.StringSlicesEqual(entriesWithBL["Note 2"], []string{"has a", "note 4"}) {
		t.Errorf("Expected %s, got %s", []string{"has a", "Note 4"}, entriesWithBL["Note 2"])
	}
	if !util.StringSlicesEqual(entriesWithBL["Note 3"], []string{}) {
		t.Errorf("Expected %s, got %s", []string{}, entriesWithBL["Note 1"])
	}
}
