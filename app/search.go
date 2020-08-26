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
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/analyzer/standard"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search/query"
	"memory/app/config"
	"memory/app/model"
	"memory/app/persist"
	"memory/util"
	"strings"

	"github.com/blevesearch/bleve"
)

var searchIndex bleve.Index

// getIndexMapping returns the default index settings for
// new and existing search indexes.
func getIndexMapping() mapping.IndexMapping {
	entryMapping := bleve.NewDocumentMapping()
	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = en.AnalyzerName
	storeOnlyFieldMapping := bleve.NewTextFieldMapping()
	storeOnlyFieldMapping.Index = false
	boolFieldMapping := bleve.NewBooleanFieldMapping()
	boolFieldMapping.Store = false
	timeMapping := bleve.NewDateTimeFieldMapping()
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Type = "text"
	keywordFieldMapping.Analyzer = standard.Name
	startFieldMapping := bleve.NewTextFieldMapping()
	startFieldMapping.Type = "text"
	startFieldMapping.Analyzer = keyword.Name
	startFieldMapping.Name = "Start"
	entryMapping.AddFieldMapping(startFieldMapping)
	entryMapping.AddFieldMappingsAt("Name", englishTextFieldMapping)
	entryMapping.AddFieldMappingsAt("Description", englishTextFieldMapping)
	entryMapping.AddFieldMappingsAt("Tags", keywordFieldMapping)
	entryMapping.AddFieldMappingsAt("EntryType", englishTextFieldMapping)
	entryMapping.AddFieldMappingsAt("Exclude", boolFieldMapping)
	entryMapping.AddFieldMappingsAt("LinksTo", keywordFieldMapping)
	entryMapping.AddFieldMappingsAt("LinkedFrom", keywordFieldMapping)
	//entryMapping.AddFieldMappingsAt("Start", keywordFieldMapping)
	entryMapping.AddFieldMappingsAt("End", keywordFieldMapping)
	entryMapping.AddFieldMappingsAt("Address", englishTextFieldMapping)
	entryMapping.AddFieldMappingsAt("Custom", englishTextFieldMapping)
	entryMapping.AddFieldMappingsAt("Modified", timeMapping)
	//TODO: Index lat/long; create/mod date
	mapping := bleve.NewIndexMapping()
	mapping.AddDocumentMapping("Entry", entryMapping)
	return mapping
}

// initSearch should be called to setup search on application
// startup after entries are loaded/available.
func initSearch() error {
	indexPath := config.SearchPath()
	if util.PathExists(indexPath) {
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

// IndexEntry adds or updates an entry in the index
func IndexEntry(entry model.Entry) error {
	return searchIndex.Index(entry.Slug(), entry)
}

// RemoveFromIndex removes an entry from the index
func RemoveFromIndex(slug string) error {
	return searchIndex.Delete(slug)
}

// RebuildSearchIndex creates a new search index of current entries.
func RebuildSearchIndex() error {
	if err := util.DelTree(config.SearchPath()); err != nil {
		return err
	}
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
		entry, err := GetEntryFromStorage(slug)
		if err != nil {
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
	req := bleve.NewSearchRequestOptions(q, util.MaxInt32, 0, false)
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
func GetEntryFromIndex(slug string) (model.Entry, bool) {
	doc, err := searchIndex.Document(slug)
	if err != nil || doc == nil {
		return model.Entry{}, false
	}
	entry := model.NewEntry("", "", "", []string{})
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
		case "Modified":
			//TODO: Suggest an example of this at https://blevesearch.com/docs/Index-Mapping/
			df, ok := field.(*document.DateTimeField)
			if ok {
				dt, err := df.DateTime()
				if err == nil {
					entry.Modified = dt
				}
			}
			//entry.Modified, _ = time.Parse("2017-08-31 00:00:00 +0000 UTC", string(field.Value()))
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
func SearchEntries(types model.EntryTypes, search string, onlyTags []string,
	anyTags []string, sort SortOrder, pageNo int, pageSize int) (EntryResults, error) {
	query := buildSearchQuery(types, search, onlyTags, anyTags)
	req := bleve.NewSearchRequestOptions(query, pageSize, (pageNo-1)*pageSize, false)
	if sort == SortName {
		req.SortBy([]string{"Name"})
	} else if sort == SortRecent {
		req.SortBy([]string{"-Modified"})
	} else {
		req.SortBy([]string{"-_score"})
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
		Sort: sort, PageNo: pageNo, PageSize: pageSize, Total: searchResult.Total, Entries: []model.Entry{}}
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

func buildSearchQuery(types model.EntryTypes, search string, onlyTags []string, anyTags []string) *query.BooleanQuery {
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
		boolQ := bleve.NewBooleanQuery()
		qname := bleve.NewMatchQuery(search)
		qname.SetField("Name")
		qname.SetBoost(3)
		otherQ := bleve.NewMatchQuery(search)
		boolQ.AddShould(qname)
		boolQ.AddShould(otherQ)
		boolQuery.AddMust(boolQ)
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

// Timeline performs a search based on start and end attributes
func Timeline(start string, end string) ([]model.Entry, error) {
	ret := []model.Entry{}
	boolQuery := bleve.NewBooleanQuery()
	if start != "" {
		startQ := bleve.NewTermRangeQuery(start, "")
		startQ.SetField("Start")
		boolQuery.AddMust(startQ)
	}
	if end != "" {
		endQ := bleve.NewTermRangeQuery("", end)
		endQ.SetField("End")
		boolQuery.AddMust(endQ)
	}
	req := bleve.NewSearchRequestOptions(boolQuery, util.MaxInt32, 0, false)
	result, err := searchIndex.Search(req)
	if err != nil {
		return ret, err
	}
	hits := result.Hits
	for _, hit := range hits {
		entry, _ := GetEntryFromIndex(hit.ID)
		ret = append(ret, entry)
	}
	return ret, nil
}
