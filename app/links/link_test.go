/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package links

import (
	"io/ioutil"
	"memory/app"
	"memory/app/model"
	"memory/util"
	"testing"
)

func TestParseLinks(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test_parse_links")
	defer util.DelTree(tempDir)
	if err != nil {
		t.Error(err)
		return
	}
	memApp, err := app.Init(tempDir)
	if err != nil {
		t.Error(err)
	}
	n1 := model.NewEntry(model.EntryTypeNote, "Exists", "", []string{})
	memApp.PutEntry(n1)
	n2 := model.NewEntry(model.EntryTypeNote, "Exists 2", "", []string{})
	memApp.PutEntry(n2)
	testParseLinks(t, memApp, 1, "[Exists]", "[Exists]", []string{"exists"})
	testParseLinks(t, memApp, 2, "text [Exists]", "text [Exists]", []string{"exists"})
	testParseLinks(t, memApp, 3, "[Exists] text", "[Exists] text", []string{"exists"})
	// we record links to pages that don't exist so they can be listed as broken links
	testParseLinks(t, memApp, 4, "[Not Exists]", "[?Not Exists]", []string{"not-exists"})
	testParseLinks(t, memApp, 5, "[Exists] [Exists  \n2]", "[Exists] [Exists  \n2]", []string{"exists", "exists-2"})
	testParseLinks(t, memApp, 6, "", "", []string{})
	testParseLinks(t, memApp, 7, "[Exists]\n[Exists 2]", "[Exists]\n[Exists 2]", []string{"exists", "exists-2"})
	testParseLinks(t, memApp, 8, "[~Exists]", "[~Exists]", []string{})
	testParseLinks(t, memApp, 9, "[?Exists]", "[Exists]", []string{"exists"})
	testParseLinks(t, memApp, 10, "[?Not Exists]", "[?Not Exists]", []string{"not-exists"})
	testParseLinks(t, memApp, 11, "[~Not Exists]", "[~Not Exists]", []string{})
	testParseLinks(t, memApp, 12, "[Exists 2]\n[Exists]", "[Exists 2]\n[Exists]", []string{"exists-2", "exists"})
	testParseLinks(t, memApp, 13, "[Exists](external)", "[Exists](external)", []string{})
}

func testParseLinks(t *testing.T, memApp *app.Memory, testNo int, input string, parsedExpected string, linksExpected []string) {
	links := ExtractLinks(input)
	parsed := RenderLinks(input, memApp.EntryExists)
	if parsed != parsedExpected {
		t.Errorf("#%d Expected parsed '%s', got '%s'", testNo, parsedExpected, parsed)
	}
	if !util.StringSlicesEqual(linksExpected, links) {
		t.Errorf("#%d Expected links '%s', got '%s'", testNo, linksExpected, links)
	}
}

//func TestResolveLinks(t *testing.T) {
//	tempDir, err := ioutil.TempDir("", "test_resolve_links")
//	defer util.DelTree(tempDir)
//	if err != nil {
//		t.Error(err)
//		return
//	}
//	memApp, err := app.Init(tempDir)
//	n1 := model.NewEntry(model.EntryTypeNote, "Exists", "", []string{})
//	memApp.PutEntry(n1)
//	n2 := model.NewEntry(model.EntryTypeNote, "Exists 2", "", []string{})
//	memApp.PutEntry(n2)
//	links := []string{"exists", "not-exists", "exists-2"}
//	resolved := ResolveLinks(links)
//	if len(resolved) != 2 {
//		t.Error("Expected len of 2, got", len(resolved))
//	}
//	if resolved[1].Name != "Exists 2" {
//		t.Error("Expected 'Exists 2', got", resolved[1].Name)
//	}
//}

func TestPopulateLinks(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test_populate_links")
	defer util.DelTree(tempDir)
	if err != nil {
		t.Error(err)
		return
	}
	memApp, err := app.Init(tempDir)
	if err != nil {
		t.Error(err)
	}
	nA := model.NewEntry(model.EntryTypeNote, "Note 1", "This note has a link to [Note 2].", []string{})
	nB := model.NewEntry(model.EntryTypeNote, "Note 2", "This note has a link to [Note 3] and [Note 2].", []string{})
	nC := model.NewEntry(model.EntryTypeNote, "Note 3", "This note has no links.", []string{})
	memApp.PutEntry(nA)
	memApp.PutEntry(nB)
	memApp.PutEntry(nC)
	n1, _ := memApp.GetEntry(util.GetSlug("Note 1"))
	n2, _ := memApp.GetEntry(util.GetSlug("Note 2"))
	n3, _ := memApp.GetEntry(util.GetSlug("Note 3"))
	// test linksTo
	links, _ := memApp.Search.Links(n1.Slug())
	if !util.StringSlicesEqual(links, []string{"note-2"}) {
		t.Error("Expected n1.LinksTo==['note-2'], got", links)
	}
	links, _ = memApp.Search.Links(n2.Slug())
	if !util.StringSlicesEqual(links, []string{"note-3", "note-2"}) {
		t.Error("Expected n2.LinksTo==['note-3','note-2'], got", links)
	}
	links, _ = memApp.Search.Links(n3.Slug())
	if !util.StringSlicesEqual(links, []string{}) {
		t.Error("Expected n3.LinksTo==[], got", links)
	}
	// test linkedFrom
	links, _ = memApp.Search.ReverseLinks(n1.Slug())
	if !util.StringSlicesEqual(links, []string{}) {
		t.Error("Expected n1.LinkedFrom==[], got", links)
	}
	links, _ = memApp.Search.ReverseLinks(n2.Slug())
	if !util.StringSlicesEqual(links, []string{"note-1", "note-2"}) {
		t.Error("Expected n2.LinkedFrom==['note-1','note-2'], got", links)
	}
	links, _ = memApp.Search.ReverseLinks(n1.Slug())
	if !util.StringSlicesEqual(links, []string{"note-2"}) {
		t.Error("Expected n3.LinkedFrom==['note-2'], got", links)
	}
}

func TestBrokenLinks(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test_broken_links")
	defer util.DelTree(tempDir)
	if err != nil {
		t.Error(err)
		return
	}
	memApp, err := app.Init(tempDir)
	if err != nil {
		t.Error(err)
	}
	nA := model.NewEntry(model.EntryTypeNote, "Note 1", "This note has a link to [Note A].", []string{})
	nB := model.NewEntry(model.EntryTypeNote, "Note 2", "This note [has a] link to [note 4] and [Note 1].", []string{})
	nC := model.NewEntry(model.EntryTypeNote, "Note 3", "This note has no links.", []string{})
	memApp.PutEntry(nA)
	memApp.PutEntry(nB)
	memApp.PutEntry(nC)
	entriesWithBL, err := memApp.Search.BrokenLinks()
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
