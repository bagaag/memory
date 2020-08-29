/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/* Searcher implementation using the go-native Bleve search engine. */

package impl

import (
	"errors"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/analyzer/standard"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search/query"
	"memory/app/config"
	"memory/app/localfs"
	"memory/app/model"
	"memory/app/persist"
	"memory/app/search"
	"memory/util"
	"strings"
)

type BleveSearch struct {
	persister   persist.Persister
	indexDir    string
	searchIndex bleve.Index
}

type BleveSearchConfig struct {
	IndexDir  string
	Persister persist.Persister
}

func NewBleveSearch(cfg BleveSearchConfig) (BleveSearch, error) {
	b := BleveSearch{persister: cfg.Persister, indexDir: cfg.IndexDir}
	return b, b.initSearch()
}

// indexMapping returns the default index settings for
// new and existing search indexes.
func (b *BleveSearch) indexMapping() mapping.IndexMapping {
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
func (b *BleveSearch) initSearch() error {
	indexPath := config.SearchPath()
	if localfs.PathExists(indexPath) {
		// open existing search index
		var err error
		b.searchIndex, err = bleve.Open(indexPath)
		if err != nil {
			return err
		}
	} else {
		if err := b.RebuildSearchIndex(); err != nil {
			return err
		}
	}
	return nil
}

// IndexEntry adds or updates an entry in the index
func (b *BleveSearch) IndexEntry(entry model.Entry) error {
	return b.searchIndex.Index(entry.Slug(), entry)
}

// RemoveFromIndex removes an entry from the index
func (b *BleveSearch) RemoveFromIndex(slug string) error {
	return b.searchIndex.Delete(slug)
}

// RebuildSearchIndex creates a new search index of current entries.
func (b *BleveSearch) RebuildSearchIndex() error {
	if err := util.DelTree(config.SearchPath()); err != nil {
		return err
	}
	// create new search index
	var err error
	b.searchIndex, err = bleve.New(config.SearchPath(), b.indexMapping())
	if err != nil {
		return err
	}
	fmt.Println("Indexing entries for search...")
	count := 0
	slugs, err := b.persister.EntrySlugs()
	if err != nil {
		return err
	}
	for _, slug := range slugs {
		entry, err := b.persister.ReadEntry(slug)
		if err != nil {
			fmt.Println("Error reading", slug, err)
			continue
		}
		if err := b.searchIndex.Index(slug, entry); err != nil {
			fmt.Println("Error indexing:", err)
		} else {
			count = count + 1
		}
	}
	fmt.Printf("Indexed %d out of %d entries.\n", count, len(slugs))
	return nil
}

// IndexedSlugs returns a slice of slugs representing entries indexed for search.
func (b *BleveSearch) IndexedSlugs() ([]string, error) {
	q := bleve.NewMatchAllQuery()
	req := bleve.NewSearchRequestOptions(q, util.MaxInt32, 0, false)
	result, err := b.searchIndex.Search(req)
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
func (b *BleveSearch) GetEntry(slug string) (model.Entry, error) {
	doc, err := b.searchIndex.Document(slug)
	if err != nil || doc == nil {
		return model.Entry{}, err
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
			//TODO: figure out why bleve stores and returns a date value from a text mapped field containing 'yyyy-mm-dd' value
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
	return entry, nil
}

// IndexedCount returns the total number of entries in the search index.
func (b *BleveSearch) IndexedCount() uint64 {
	i, _ := b.searchIndex.DocCount()
	return i
}

// SearchEntries returns a page of results based on multiple filters and search query.
func (b *BleveSearch) SearchEntries(types model.EntryTypes, keywords string, onlyTags []string,
	anyTags []string, sort search.SortOrder, pageNo int, pageSize int) (search.EntryResults, error) {
	query := b.buildSearchQuery(types, keywords, onlyTags, anyTags)
	req := bleve.NewSearchRequestOptions(query, pageSize, (pageNo-1)*pageSize, false)
	if sort == search.SortName {
		req.SortBy([]string{"Name"})
	} else if sort == search.SortRecent {
		req.SortBy([]string{"-Modified"})
	} else {
		req.SortBy([]string{"-_score"})
	}
	searchResult, err := b.searchIndex.Search(req)
	if err != nil {
		return search.EntryResults{}, err
	}
	ids := []string{}
	for _, hit := range searchResult.Hits {
		ids = append(ids, hit.ID)
	}
	results := search.EntryResults{Types: types, Search: keywords, AnyTags: anyTags, OnlyTags: onlyTags,
		Sort: sort, PageNo: pageNo, PageSize: pageSize, Total: searchResult.Total, Entries: []model.Entry{}}
	for _, id := range ids {
		entry, err := b.GetEntry(id)
		if err != nil {
			if _, notFound := err.(model.EntryNotFound); notFound {
				return search.EntryResults{}, errors.New("Document in search results not found in index: " + id)
			} else {
				return search.EntryResults{}, err
			}
		}
		results.Entries = append(results.Entries, entry)
	}
	return results, nil
}

// RefreshResults re-runs a search to freshen the results in case any entries have been modified.
func (b *BleveSearch) RefreshResults(stale search.EntryResults) (search.EntryResults, error) {
	return b.SearchEntries(stale.Types, stale.Search, stale.OnlyTags, stale.AnyTags, stale.Sort, stale.PageNo, stale.PageSize)
}

func (b *BleveSearch) buildSearchQuery(types model.EntryTypes, keywords string, onlyTags []string, anyTags []string) *query.BooleanQuery {
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
	if keywords != "" {
		boolQ := bleve.NewBooleanQuery()
		qname := bleve.NewMatchQuery(keywords)
		qname.SetField("Name")
		qname.SetBoost(3)
		otherQ := bleve.NewMatchQuery(keywords)
		boolQ.AddShould(qname)
		boolQ.AddShould(otherQ)
		boolQuery.AddMust(boolQ)
	}
	// add "get all" query if no other queries are being applied
	if types.HasAll() && len(anyTags) == 0 && len(onlyTags) == 0 && keywords == "" {
		all := bleve.NewMatchAllQuery()
		boolQuery.AddMust(all)
	}
	return boolQuery
}

// EntryCount returns the total number of entries in the index.
func (b *BleveSearch) EntryCount() uint64 {
	c, _ := b.searchIndex.DocCount()
	return c
}

// Timeline performs a search based on start and end attributes
func (b *BleveSearch) Timeline(start string, end string) ([]model.Entry, error) {
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
	result, err := b.searchIndex.Search(req)
	if err != nil {
		return ret, err
	}
	hits := result.Hits
	for _, hit := range hits {
		entry, _ := b.GetEntry(hit.ID)
		ret = append(ret, entry)
	}
	return ret, nil
}
