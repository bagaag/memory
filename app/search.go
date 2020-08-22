/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/*
This file contains functions associated with the search feature, currently
backed by https://blevesearch.com
*/

package app

import (
	"errors"
	"fmt"
	"memory/app/config"
	"memory/app/persist"
	"strings"

	//"github.com/blevesearch/bleve/analysis/analyzer/keyword"

	"github.com/blevesearch/bleve/analysis/analyzer/standard"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search/query"

	"github.com/blevesearch/bleve"
)

var searchIndex bleve.Index

// BleveType implements the alternate bleve.Classifier interface to avoid a
// naming conflict with .Type.
func (entry *Entry) BleveType() string {
	return "Entry"
}

// getIndexMapping returns the default index settings for
// new and existing search indexes.
func getIndexMapping() mapping.IndexMapping {
	entryMapping := bleve.NewDocumentMapping()
	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Store = true
	englishTextFieldMapping.Index = true
	englishTextFieldMapping.Analyzer = en.AnalyzerName
	storeOnlyFieldMapping := bleve.NewTextFieldMapping()
	storeOnlyFieldMapping.Store = true
	storeOnlyFieldMapping.Index = false
	boolFieldMapping := bleve.NewBooleanFieldMapping()
	boolFieldMapping.Store = false
	boolFieldMapping.Index = true
	//keywordAnalyzer :=
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Store = true
	keywordFieldMapping.Index = true
	keywordFieldMapping.Analyzer = standard.Name
	entryMapping.AddFieldMappingsAt("Name", englishTextFieldMapping)
	entryMapping.AddFieldMappingsAt("Description", englishTextFieldMapping)
	entryMapping.AddFieldMappingsAt("Tags", keywordFieldMapping)
	entryMapping.AddFieldMappingsAt("EntryType", englishTextFieldMapping)
	entryMapping.AddFieldMappingsAt("Exclude", boolFieldMapping)
	entryMapping.AddFieldMappingsAt("LinksTo", keywordFieldMapping)
	entryMapping.AddFieldMappingsAt("LinkedFrom", keywordFieldMapping)
	entryMapping.AddFieldMappingsAt("Start", keywordFieldMapping)
	entryMapping.AddFieldMappingsAt("End", keywordFieldMapping)
	entryMapping.AddFieldMappingsAt("Address", englishTextFieldMapping)
	entryMapping.AddFieldMappingsAt("Custom", englishTextFieldMapping)
	//TODO: Index lat/long; create/mod date
	mapping := bleve.NewIndexMapping()
	mapping.AddDocumentMapping("Entry", entryMapping)
	return mapping
}

// inistSearch should be called to setup search on application
// startup after entries are loaded/available.
func initSearch() error {
	indexPath := config.SearchPath()
	if persist.PathExists(indexPath) {
		// open existing search index
		var err error
		searchIndex, err = bleve.Open(indexPath)
		if err != nil {
			return err
		}
	} else {
		if err := RebuildSearchIndex(); err != nil {
			return err
		}
	}
	return nil
}

// closeSearch should be called at app shutdown
func closeSearch() error {
	return searchIndex.Close()
}

// executeSearch returns a list of entry slugs matching the given keywords.
func executeSearch(keywords string) ([]string, error) {
	query := bleve.NewQueryStringQuery(keywords)
	search := bleve.NewSearchRequest(query)
	searchResult, err := searchIndex.Search(search)
	if err != nil {
		return nil, err
	}
	keys := []string{}
	for _, hit := range searchResult.Hits {
		keys = append(keys, hit.ID)
	}
	return keys, nil
}

// IndexEntry adds or updates an entry in the index
func IndexEntry(entry Entry) error {
	return searchIndex.Index(entry.Slug(), entry)
}

// RemoveFromIndex removes an entry from the index
func RemoveFromIndex(slug string) error {
	return searchIndex.Delete(slug)
}

// RebuildSearchIndex creates a new search index of current entries.
func RebuildSearchIndex() error {
	// create new search index
	var err error
	searchIndex, err = bleve.New(config.SearchPath(), getIndexMapping())
	if err != nil {
		return err
	}
	fmt.Println("Indexing entries for search...")
	count := 0
	slugs, err := persist.EntrySlugs()
	if err != nil {
		return err
	}
	for _, slug := range slugs {
		entry, exists, err := GetEntryFromStorage(slug)
		if !exists {
			fmt.Println("Error:", slug, "listed from storage but not found")
			continue
		} else if err != nil {
			fmt.Println("Error reading", slug, err)
			continue
		}
		if err := searchIndex.Index(slug, entry); err != nil {
			fmt.Println("Error indexing:", err)
		} else {
			count = count + 1
		}
	}
	fmt.Println("Parsing links...")
	populateLinks()
	fmt.Printf("Indexed %d out of %d entries.\n", count, len(slugs))
	return nil
}

// IndexedSlugs returns a slice of slugs representing entries indexed for search.
func IndexedSlugs() ([]string, error) {
	q := bleve.NewMatchAllQuery()
	req := bleve.NewSearchRequest(q)
	result, err := searchIndex.Search(req)
	if err != nil {
		return nil, err
	}
	slugs := []string{}
	for _, hit := range result.Hits {
		slugs = append(slugs, hit.ID)
	}
	return slugs, nil
}

// GetEntryFromIndex returns an entry from the search index suitable for display.
func GetEntryFromIndex(slug string) (Entry, bool) {
	doc, err := searchIndex.Document(slug)
	if err != nil || doc == nil {
		return Entry{}, false
	}
	entry := NewEntry("", "", "", []string{})
	for _, field := range doc.Fields {
		switch field.Name() {
		case "Name":
			entry.Name = string(field.Value())
		case "Description":
			entry.Description = string(field.Value())
		case "EntryType":
			entry.Type = string(field.Value())
		case "Tags": // there's a separate Tags field for each tag value in a document
			entry.Tags = append(entry.Tags, string(field.Value()))
		case "LinksTo":
			entry.LinksTo = append(entry.LinksTo, string(field.Value()))
		case "LinkedFrom":
			entry.LinkedFrom = append(entry.LinkedFrom, string(field.Value()))
		case "Start":
			entry.Start = string(field.Value())
		case "End":
			entry.End = string(field.Value())
		case "Address":
			entry.Address = string(field.Value())
		default:
			if strings.HasPrefix(field.Name(), "Custom.") {
				key := strings.Split(field.Name(), ".")[1]
				entry.Custom[key] = string(field.Value())
			}
		}
	}
	return entry, true
}

// IndexedCount returns the total number of entries in the search index.
func IndexedCount() uint64 {
	i, _ := searchIndex.DocCount()
	return i
}

// SearchEntries returns a page of results based on multiple filters and search query.
func SearchEntries(types EntryTypes, search string, onlyTags []string,
	anyTags []string, sort SortOrder, pageNo int, pageSize int) (EntryResults, error) {
	query := buildSearchQuery(types, search, onlyTags, anyTags)
	req := bleve.NewSearchRequestOptions(query, pageSize, (pageNo-1)*pageSize, false)
	if sort == SortName {
		req.SortBy([]string{"Name"})
	} else if sort == SortRecent {
		req.SortBy([]string{"Modified"})
	} else {
		req.SortBy([]string{"_score"})
	}
	searchResult, err := searchIndex.Search(req)
	if err != nil {
		return EntryResults{}, err
	}
	ids := []string{}
	for _, hit := range searchResult.Hits {
		ids = append(ids, hit.ID)
	}
	results := EntryResults{Types: types, Search: search, AnyTags: anyTags, OnlyTags: onlyTags,
		Sort: sort, PageNo: pageNo, PageSize: pageSize, Total: searchResult.Total, Entries: []Entry{}}
	for _, id := range ids {
		entry, exists := GetEntryFromIndex(id)
		if !exists {
			return EntryResults{}, errors.New("Document in search results not found in index: " + id)
		}
		results.Entries = append(results.Entries, entry)
	}
	return results, nil
}

// RefreshResults re-runs a search to freshen the results in case any entries have been modified.
func RefreshResults(stale EntryResults) (EntryResults, error) {
	return SearchEntries(stale.Types, stale.Search, stale.OnlyTags, stale.AnyTags, stale.Sort, stale.PageNo, stale.PageSize)
}

func buildSearchQuery(types EntryTypes, search string, onlyTags []string, anyTags []string) *query.BooleanQuery {
	boolQuery := bleve.NewBooleanQuery()
	// process types
	if !types.HasAll() {
		typeQuery := bleve.NewBooleanQuery()
		if types.Event {
			q := bleve.NewMatchQuery("Event")
			q.FieldVal = "EntryType"
			typeQuery.AddShould(q)
		}
		if types.Person {
			q := bleve.NewMatchQuery("Person")
			q.FieldVal = "EntryType"
			typeQuery.AddShould(q)
		}
		if types.Place {
			q := bleve.NewMatchQuery("Place")
			q.FieldVal = "EntryType"
			typeQuery.AddShould(q)
		}
		if types.Thing {
			q := bleve.NewMatchQuery("Thing")
			q.FieldVal = "EntryType"
			typeQuery.AddShould(q)
		}
		if types.Note {
			q := bleve.NewMatchQuery("Note")
			q.FieldVal = "EntryType"
			typeQuery.AddShould(q)
		}
		typeQuery.SetMinShould(1)
		boolQuery.AddMust(typeQuery)
	}
	// any tags
	if len(anyTags) > 0 {
		tagsQuery := bleve.NewBooleanQuery()
		for _, tag := range anyTags {
			tagQuery := bleve.NewMatchPhraseQuery(tag)
			tagQuery.SetField("Tags")
			tagsQuery.AddShould(tagQuery)
		}
		tagsQuery.SetMinShould(1)
		boolQuery.AddMust(tagsQuery)
	}
	// only tags (all results must have all these tags)
	if len(onlyTags) > 0 {
		tagsQuery := bleve.NewBooleanQuery()
		for _, tag := range onlyTags {
			tagQuery := bleve.NewMatchPhraseQuery(tag)
			tagQuery.SetField("Tags")
			tagsQuery.AddMust(tagQuery)
		}
		boolQuery.AddMust(tagsQuery)
	}
	// add keyword search
	if search != "" {
		q := bleve.NewQueryStringQuery(search)
		boolQuery.AddMust(q)
	}
	// add "get all" query if no other queries are being applied
	if types.HasAll() && len(anyTags) == 0 && len(onlyTags) == 0 && search == "" {
		all := bleve.NewMatchAllQuery()
		boolQuery.AddMust(all)
	}
	return boolQuery
}

// EntryCount returns the total number of entries in the index.
func EntryCount() uint64 {
	c, _ := searchIndex.DocCount()
	return c
}
