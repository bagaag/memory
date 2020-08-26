/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package app

import (
	"fmt"
	"io/ioutil"
	"log"
	"memory/app/model"
	"memory/util"
	"testing"

	"github.com/blevesearch/bleve"
)

/* This file contains functions to support full text entry search. */

var setup1 = func(t *testing.T) func(t *testing.T) {
	home, err := ioutil.TempDir("", "search_test_setup1")
	if err = initSearch(); err != nil {
		t.Error(err)
	}
	Init(home)
	e1 := model.NewEntry(model.EntryTypeNote, "Apple Heresay", "Yours is no disgrace.", []string{"tag1", "tag0"})
	e2 := model.NewEntry(model.EntryTypeNote, "Bungled Apple", "Shaky groove turtle.", []string{"tag2", "tag1"})
	e3 := model.NewEntry(model.EntryTypeEvent, "Frenetic Plum", "Undersea groove turntable swing.", []string{"tag3"})
	e3.Start = "2020"
	PutEntry(e1)
	PutEntry(e2)
	PutEntry(e3)
	Save()
	return func(t *testing.T) {
		log.Println("Deleting", home)
		util.DelTree(home)
	}
}
var setup2 = func(t *testing.T) func(t *testing.T) {
	home, err := ioutil.TempDir("", "search_test_setup1")
	if err = initSearch(); err != nil {
		t.Error(err)
	}
	Init(home)
	e1 := model.NewEntry(model.EntryTypeNote, "Apple Heresay", "Yours is no disgrace.", []string{"tag1", "tag0"})
	e2 := model.NewEntry(model.EntryTypeNote, "Bungled Apple", "Shaky groove turtle.", []string{"tag2", "tag1"})
	e3 := model.NewEntry(model.EntryTypeEvent, "Frenetic Plum", "Undersea groove turntable swing.", []string{"tag3"})
	e3.Start = "2020"
	e4 := model.NewEntry(model.EntryTypeEvent, "Links To e1", "A peopled [Apple Heresay].", []string{"groove turtle"})
	e4.Start = "2020"
	PutEntry(e1)
	PutEntry(e2)
	PutEntry(e3)
	PutEntry(e4)
	populateLinks()
	Save()
	return func(t *testing.T) {
		log.Println("Deleting", home)
		util.DelTree(home)
	}
}

func TestLinksToSearch(t *testing.T) {
	teardown2 := setup2(t)
	defer teardown2(t)
	e4, exists := GetEntryFromIndex(util.GetSlug("Links to e1"))
	if !exists {
		t.Error("e4 doesn't exist")
	}
	if len(e4.LinksTo) < 1 {
		t.Error("e4 has no linksto")
	}
	if e4.LinksTo[0] != "apple-heresay" {
		t.Error("Expected 'apple-heresay', got", e4.LinksTo[0])
	}
	query := bleve.NewMatchPhraseQuery("apple-heresay")
	query.SetField("LinksTo")
	search := bleve.NewSearchRequest(query)
	searchResult, err := searchIndex.Search(search)
	if err != nil {
		t.Error(err)
	}
	if len(searchResult.Hits) != 1 {
		t.Error("Expected 1 result, got", len(searchResult.Hits))
	}
	for _, hit := range searchResult.Hits {
		fmt.Println(hit.ID)
	}
}

func TestTagsSearch(t *testing.T) {
	teardown2 := setup2(t)
	defer teardown2(t)
	query := bleve.NewMatchPhraseQuery("groove-turtle")
	query.SetField("Tags")
	search := bleve.NewSearchRequest(query)
	searchResult, err := searchIndex.Search(search)
	if err != nil {
		t.Error(err)
	}
	if len(searchResult.Hits) != 1 {
		t.Error("Expected 1 result, got", len(searchResult.Hits))
	}
	for _, hit := range searchResult.Hits {
		fmt.Println(hit.ID)
	}
}

func TestSearch(t *testing.T) {
	teardown1 := setup1(t)
	defer teardown1(t)
	// document test
	searchDocumentTest(t, 1)
	// type test
	searchTypeTest(t, 2, "EntryType:Event", []string{"frenetic-plum"})
	// entry search
	searchEntriesTest(t, 3)
	// entry paging
	searchEntriesPagingTest(t, 20)
}

func searchEntriesPagingTest(t *testing.T, num int) {
	// page 1 of 2
	results, err := SearchEntries(model.EntryTypes{}, "", []string{}, []string{}, SortName, 1, 2)
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
	results, err = SearchEntries(model.EntryTypes{}, "", []string{}, []string{}, SortName, 2, 2)
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

func searchEntriesTest(t *testing.T, num int) {
	// all entries of type Note and Event
	results, err := SearchEntries(model.EntryTypes{Note: true, Event: true}, "", []string{}, []string{}, SortScore, 1, 10)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 3 {
		t.Errorf("%d. Expected 3, got %d", num, len(results.Entries))
	}
	num = num + 1
	// only Note entries
	results, err = SearchEntries(model.EntryTypes{Note: true}, "", []string{}, []string{}, SortScore, 1, 10)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 2 {
		t.Errorf("%d. Expected 2, got %d", num, len(results.Entries))
	}
	num = num + 1
	// Note entries containing apple
	results, err = SearchEntries(model.EntryTypes{Note: true}, "apple", []string{}, []string{}, SortScore, 1, 10)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 2 {
		t.Errorf("%d. Expected 2, got %d", num, len(results.Entries))
	}
	num = num + 1
	// Any type of entries containing apple
	results, err = SearchEntries(model.EntryTypes{Note: true}, "apple", []string{}, []string{}, SortScore, 1, 10)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 2 {
		t.Errorf("%d. Expected 2, got %d", num, len(results.Entries))
	}
	num = num + 1
	// Entries containing apple with tag2
	results, err = SearchEntries(model.EntryTypes{Note: true}, "apple", []string{"tag2"}, []string{}, SortScore, 1, 10)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 1 {
		t.Errorf("%d. Expected 1, got %d", num, len(results.Entries))
	}
	num = num + 1
	// Entries with tag0 AND tag1
	results, err = SearchEntries(model.EntryTypes{}, "", []string{"tag0", "tag1"}, []string{}, SortScore, 1, 10)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 1 {
		t.Errorf("%d. Expected 1, got %d", num, len(results.Entries))
	}
	num = num + 1
	// Entries with tag0 or tag1
	results, err = SearchEntries(model.EntryTypes{}, "", []string{}, []string{"tag0", "tag1"}, SortScore, 1, 10)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 2 {
		t.Errorf("%d. Expected 2, got %d", num, len(results.Entries))
	}
	num = num + 1
	// Get All entries
	results, err = SearchEntries(model.EntryTypes{}, "", []string{}, []string{}, SortScore, 1, 10)
	if err != nil {
		t.Error(num, err)
	}
	if len(results.Entries) != 3 {
		t.Errorf("%d. Expected 3, got %d", num, len(results.Entries))
	}
}

func searchTypeTest(t *testing.T, num int, keywords string, expect []string) {
	query := bleve.NewQueryStringQuery(keywords)
	//query.SetField("EntryType")
	search := bleve.NewSearchRequest(query)
	searchResult, err := searchIndex.Search(search)
	if err != nil {
		t.Error(num, err)
	}
	r := []string{}
	for _, hit := range searchResult.Hits {
		r = append(r, hit.ID)
	}
	if err != nil {
		t.Error(num, err)
	}
	if !util.StringSlicesEqual(r, expect) {
		t.Errorf("%d. Expected %v, got %v", num, expect, r)
	}
}

func searchDocumentTest(t *testing.T, num int) {
	// get doc from index
	entry, exists := GetEntryFromIndex("apple-heresay")
	if !exists {
		t.Error(num, "apple-heresay doesn't exist in index, but should")
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