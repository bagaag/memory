/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package test

import (
	"io/ioutil"
	"log"
	"memory/app/memory"
	"memory/app/model"
	"memory/app/search"
	"memory/util"
	"testing"
)

/* This file contains functions to support full text entry search. */

var setup1 = func(t *testing.T) (*memory.Memory, func(t *testing.T)) {
	home, err := ioutil.TempDir("", "search_test_setup1")
	if err != nil {
		t.Error(err)
	}
	memApp, err := memory.Init(home)
	if err != nil {
		t.Error(err)
	}
	e1 := model.NewEntry(model.EntryTypeNote, "Apple Heresay", "Yours is no disgrace.", []string{"tag1", "tag0"})
	e2 := model.NewEntry(model.EntryTypeNote, "Bungled Apple", "Shaky groove turtle.", []string{"tag2", "tag1"})
	e3 := model.NewEntry(model.EntryTypeEvent, "Frenetic Plum", "Undersea groove turntable swing.", []string{"tag3"})
	e3.Start = "2020"
	consumeError(t, memApp.PutEntry(e1))
	consumeError(t, memApp.PutEntry(e2))
	consumeError(t, memApp.PutEntry(e3))
	return memApp, func(t *testing.T) {
		log.Println("Deleting", home)
		consumeError(t, util.DelTree(home))
	}
}
var setup2 = func(t *testing.T) (*memory.Memory, func(t *testing.T)) {
	home, err := ioutil.TempDir("", "search_test_setup1")
	if err != nil {
		t.Error(err)
	}
	memApp, err := memory.Init(home)
	if err != nil {
		t.Error(err)
	}
	e1 := model.NewEntry(model.EntryTypeNote, "Apple Heresay", "Yours is no disgrace.", []string{"tag1", "tag0"})
	e2 := model.NewEntry(model.EntryTypeNote, "Bungled Apple", "Shaky groove turtle.", []string{"tag2", "tag1"})
	e3 := model.NewEntry(model.EntryTypeEvent, "Frenetic Plum", "Undersea groove turntable swing.", []string{"tag3"})
	e3.Start = "2020"
	e4 := model.NewEntry(model.EntryTypeEvent, "Links To e1", "A peopled [Apple Heresay].", []string{"groove turtle"})
	e4.Start = "2020"
	consumeError(t, memApp.PutEntry(e1))
	consumeError(t, memApp.PutEntry(e2))
	consumeError(t, memApp.PutEntry(e3))
	consumeError(t, memApp.PutEntry(e4))
	return memApp, func(t *testing.T) {
		log.Println("Deleting", home)
		consumeError(t, util.DelTree(home))
	}
}

func consumeError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func TestLinksToSearch(t *testing.T) {
	memApp, teardown2 := setup2(t)
	defer teardown2(t)
	slug := util.GetSlug("Links to e1")
	links, err := memApp.Search.Links(slug)
	if err != nil {
		t.Error(err)
	}
	if len(links) < 1 {
		t.Error("e4 has no linksto")
	}
	if links[0] != "apple-heresay" {
		t.Error("Expected 'apple-heresay', got", links[0])
	}
	// TODO: This should be implemented as `memApp.Search.LinksTo(slug) []Entry` if needed
	//query := bleve.NewMatchPhraseQuery("apple-heresay")
	//query.SetField("LinksTo")
	//search := bleve.NewSearchRequest(query)
	//searchResult, err := searchIndex.Search(search)
	//if err != nil {
	//	t.Error(err)
	//}
	//if len(searchResult.Hits) != 1 {
	//	t.Error("Expected 1 result, got", len(searchResult.Hits))
	//}
	//for _, hit := range searchResult.Hits {
	//	fmt.Println(hit.ID)
	//}
}

func TestTagsSearch(t *testing.T) {
	//TODO: This should be implemented as `memApp.Search.HasTag(tag) []Entry` if needed
	//memApp, teardown2 := setup2(t)
	//defer teardown2(t)
	//query := bleve.NewMatchPhraseQuery("groove-turtle")
	//query.SetField("Tags")
	//search := bleve.NewSearchRequest(query)
	//searchResult, err := searchIndex.Search(search)
	//if err != nil {
	//	t.Error(err)
	//}
	//if len(searchResult.Hits) != 1 {
	//	t.Error("Expected 1 result, got", len(searchResult.Hits))
	//}
	//for _, hit := range searchResult.Hits {
	//	fmt.Println(hit.ID)
	//}
}

func TestSearch(t *testing.T) {
	memApp, teardown1 := setup1(t)
	defer teardown1(t)
	// document test
	searchDocumentTest(t, memApp, 1)
	// type test
	//searchTypeTest(t, memApp, 2, "EntryType:Event", []string{"frenetic-plum"})
	// entry search
	searchEntriesTest(t, memApp, 3)
	// entry paging
	searchEntriesPagingTest(t, memApp, 20)
}

func searchEntriesPagingTest(t *testing.T, memApp *memory.Memory, num int) {
	// page 1 of 2
	results, err := memApp.Search.SearchEntries(model.EntryTypes{}, "", []string{}, []string{}, search.SortName, 1, 2)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 2 {
		t.Errorf("%d. Expected 2, got %d", num, len(results.Entries))
	}
	if results.Total != 3 {
		t.Errorf("%d. Expected 3, got %d", num, len(results.Entries))
	}
	num = num + 1
	// page 2 of 2
	results, err = memApp.Search.SearchEntries(model.EntryTypes{}, "", []string{}, []string{}, search.SortName, 2, 2)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 1 {
		t.Errorf("%d. Expected 1, got %d", num, len(results.Entries))
	}
	if results.Total != 3 {
		t.Errorf("%d. Expected 3, got %d", num, len(results.Entries))
	}
	num = num + 1
}

func searchEntriesTest(t *testing.T, memApp *memory.Memory, num int) {
	// all entries of type Note and Event
	results, err := memApp.Search.SearchEntries(model.EntryTypes{Note: true, Event: true}, "", []string{}, []string{}, search.SortScore, 1, 10)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 3 {
		t.Errorf("%d. Expected 3, got %d", num, len(results.Entries))
	}
	num = num + 1
	// only Note entries
	results, err = memApp.Search.SearchEntries(model.EntryTypes{Note: true}, "", []string{}, []string{}, search.SortScore, 1, 10)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 2 {
		t.Errorf("%d. Expected 2, got %d", num, len(results.Entries))
	}
	num = num + 1
	// Note entries containing apple
	results, err = memApp.Search.SearchEntries(model.EntryTypes{Note: true}, "apple", []string{}, []string{}, search.SortScore, 1, 10)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 2 {
		t.Errorf("%d. Expected 2, got %d", num, len(results.Entries))
	}
	num = num + 1
	// Any type of entries containing apple
	results, err = memApp.Search.SearchEntries(model.EntryTypes{Note: true}, "apple", []string{}, []string{}, search.SortScore, 1, 10)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 2 {
		t.Errorf("%d. Expected 2, got %d", num, len(results.Entries))
	}
	num = num + 1
	// Entries containing apple with tag2
	results, err = memApp.Search.SearchEntries(model.EntryTypes{Note: true}, "apple", []string{"tag2"}, []string{}, search.SortScore, 1, 10)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 1 {
		t.Errorf("%d. Expected 1, got %d", num, len(results.Entries))
	}
	num = num + 1
	// Entries with tag0 AND tag1
	results, err = memApp.Search.SearchEntries(model.EntryTypes{}, "", []string{"tag0", "tag1"}, []string{}, search.SortScore, 1, 10)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 1 {
		t.Errorf("%d. Expected 1, got %d", num, len(results.Entries))
	}
	num = num + 1
	// Entries with tag0 or tag1
	results, err = memApp.Search.SearchEntries(model.EntryTypes{}, "", []string{}, []string{"tag0", "tag1"}, search.SortScore, 1, 10)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 2 {
		t.Errorf("%d. Expected 2, got %d", num, len(results.Entries))
	}
	num = num + 1
	// Get All entries
	results, err = memApp.Search.SearchEntries(model.EntryTypes{}, "", []string{}, []string{}, search.SortScore, 1, 10)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 3 {
		t.Errorf("%d. Expected 3, got %d", num, len(results.Entries))
	}
}

func searchDocumentTest(t *testing.T, memApp *memory.Memory, num int) {
	// get doc from index
	entry, err := memApp.Search.Stub("apple-heresay")
	if err != nil {
		t.Error(num, "apple-heresay doesn't exist in index, but should:", err)
	}
	if entry.Name != "Apple Heresay" {
		t.Error(num, "Expected 'Apple heresay' but got", entry.Name)
	}
	if entry.Description != "Yours is no disgrace." {
		t.Error(num, "Expected 'Yours is no disgrace.' but got", entry.Description)
	}
	if entry.Type != "Note" {
		t.Error(num, "Expected 'Note', got", entry.Type)
	}
	if !util.StringSlicesEqual(entry.Tags, []string{"tag1", "tag0"}) {
		t.Error(num, "Expected 'tag1,tag0', got", entry.Tags)
	}

}
