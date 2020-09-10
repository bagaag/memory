/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package test

import (
	"fmt"
	"io/ioutil"
	"log"
	"memory/app/memory"
	"memory/app/model"
	"memory/app/search"
	"memory/util"
	"os"
	"strconv"
	"testing"
)

/* This file contains functions to support full text entry search. */

func initMemApp(t *testing.T, path string) (*memory.Memory, string) {
	home, err := ioutil.TempDir("", "search_test_setup1")
	if err != nil {
		t.Error(err)
	}
	memApp, err := memory.Init(home)
	if err != nil {
		t.Error(err)
		os.Exit(1)
	}
	return memApp, home
}

var setup1 = func(t *testing.T) (*memory.Memory, func(t *testing.T)) {
	memApp, home := initMemApp(t, "search_test_setup1")
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
	memApp, home := initMemApp(t, "search_test_setup2")
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

var setup3 = func(t *testing.T, dates [][]string) (*memory.Memory, func(t *testing.T)) {
	memApp, home := initMemApp(t, "search_test_setup3")
	testEntries := []model.Entry{}
	for i, set := range dates {
		testEntries = append(testEntries, model.Entry{
			Type:   model.EntryTypeEvent,
			Name:   "E" + strconv.Itoa(i+1),
			Tags:   []string{},
			Custom: make(map[string]string),
			Start:  set[0],
			End:    set[1],
		})
	}
	for _, entry := range testEntries {
		consumeError(t, memApp.PutEntry(entry))
	}
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

func TestDateStorage(t *testing.T) {
	dates := [][]string{
		//start date is required; end date missing should be treated
		// as if it had the same value as start date
		{"2001-03-02", "2007-05-17"},
		{"2000-01", "2010-04"},
		{"2000", "2010"},
	}
	memApp, teardown3 := setup3(t, dates)
	defer teardown3(t)
	slugs := []string{"e1", "e2", "e3"}
	for ix, slug := range slugs {
		d := dates[ix]
		e, _ := memApp.Search.Stub(slug)
		if e.Start != d[0] {
			t.Errorf("Expected Start '%s', got '%s'", d[0], e.Start)
		}
		if e.End != d[1] {
			t.Errorf("Expected End '%s', got '%s'", d[1], e.End)
		}
	}
}

func TestTimeline(t *testing.T) {
	// stores a test case definition
	type test struct {
		start    string
		end      string
		expected []string
	}
	// entry dates to test against
	dates := [][]string{
		//start date is required; end date missing should be treated
		// as if it had the same value as start date
		{"2000", ""},
		{"2000-02", ""},
		{"2001-03-01", ""},
		{"2002-03-10", "2003-01-02"},
		{"2003-03-22", "2004-02-10"},
		{"2004-01-01", "2008-01-02"},
	}
	// define test cases
	tests := []test{
		{"", "", []string{"E1", "E2", "E3", "E4", "E5", "E6"}},
		{"2001", "", []string{"E3", "E4", "E5", "E6"}},
		{"2001", "2003-01-03", []string{"E3", "E4"}},
		{"2002-02", "2004", []string{"E4", "E5"}},
		{"2002-05-01", "2009", []string{"E5", "E6"}},
	}
	// prints a result in shorthand
	printEntries := func(es []model.Entry) string {
		s := ""
		for _, e := range es {
			s = s + fmt.Sprintf("%s:[%s/%s] ", e.Name, e.Start, e.End)
		}
		return s
	}
	// tests result against expected value
	gotExpected := func(result []model.Entry, expected []string) bool {
		if len(result) != len(expected) {
			return false
		}
		for i, e := range result {
			if expected[i] != e.Name {
				return false
			}
		}
		return true
	}
	// init app and create entries
	memApp, teardown3 := setup3(t, dates)
	defer teardown3(t)
	// run test cases
	for i, testCase := range tests {
		r, e := memApp.Search.Timeline(testCase.start, testCase.end)
		testNum := strconv.Itoa(i+1) + "."
		if e != nil {
			t.Error(testNum, e)
		} else if !gotExpected(r, testCase.expected) {
			t.Error(testNum, "Expected", testCase.expected, "got", printEntries(r))
		} else {
			fmt.Println(testNum, "OK: Expected", testCase.expected, "got", printEntries(r))
		}
	}
}
