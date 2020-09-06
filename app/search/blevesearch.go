/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/* Searcher implementation using the go-native Bleve search engine. */

package search

import (
	"errors"
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/standard"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search/query"
	"memory/app/config"
	"memory/app/links"
	"memory/app/localfs"
	"memory/app/model"
	"memory/app/persist"
	"memory/util"
	"strconv"
	"strings"
	"time"
)

// BleveSearch is a search implementation based on the go-native Bleve search engine.
type BleveSearch struct {
	persister   persist.Persister
	indexDir    string
	searchIndex bleve.Index
}

// BleveSearchConfig defines the values required to create an instance of BleveSearch.
type BleveSearchConfig struct {
	IndexDir  string
	Persister persist.Persister
}

// IndexedEntry is a representation of model.Entry suited for indexing by Bleve search.
type IndexedEntry struct {
	Name           string
	Description    string
	Tags           []string
	Links          []string
	Created        time.Time
	Modified       time.Time
	EntryType      string
	Start          time.Time // Events
	StartPrecision int       // 0=Year, 1=Month, 2=Day
	End            time.Time // Events
	EndPrecision   int       // 0=Year, 1=Month, 2=Day
	Location       Location
	Address        string // Place
	Custom         map[string]string
	Exclude        bool // Supports ability to search for all entries
}

type Location struct {
	Lat float64
	Lon float64
}

func NewBleveSearch(cfg BleveSearchConfig) (BleveSearch, error) {
	b := BleveSearch{persister: cfg.Persister, indexDir: cfg.IndexDir}
	return b, b.initSearch()
}

// NewIndexedEntry converts a model.Entry to an IndexedEntry.
func NewIndexedEntry(entry model.Entry) IndexedEntry {
	indexed := IndexedEntry{
		Name:        entry.Name,
		Description: util.TruncateAtWhitespace(entry.Description, 200),
		Tags:        entry.Tags,
		Links:       links.ExtractLinks(entry.Description),
		Created:     entry.Created,
		Modified:    entry.Modified,
		EntryType:   entry.Type,
		Address:     entry.Address,
		Custom:      entry.Custom,
		Exclude:     false,
	}
	if entry.Start != "" {
		date, precision := parseFlexDate(entry.Start)
		indexed.Start = date
		indexed.StartPrecision = precision
	}
	if entry.End != "" {
		date, precision := parseFlexDate(entry.End)
		indexed.End = date
		indexed.EndPrecision = precision
	}
	if entry.Latitude != "" && entry.Longitude != "" {
		lat, err1 := strconv.ParseFloat(entry.Latitude, 64)
		lon, err2 := strconv.ParseFloat(entry.Longitude, 64)
		if err1 != nil && err2 != nil {
			indexed.Location = Location{lat, lon}
		}
	}
	if indexed.Custom == nil {
		indexed.Custom = make(map[string]string)
	}
	return indexed
}

func (ix *IndexedEntry) Entry() model.Entry {
	entry := model.Entry{
		Name:        ix.Name,
		Description: ix.Description,
		Tags:        ix.Tags,
		Start:       flexDate(ix.Start, ix.StartPrecision),
		End:         flexDate(ix.End, ix.EndPrecision),
		Created:     ix.Created,
		Modified:    ix.Modified,
		Type:        ix.EntryType,
		Address:     ix.Address,
		Custom:      ix.Custom,
	}
	if ix.Location.Lat > 0 {
		entry.Latitude = strconv.FormatFloat(ix.Location.Lat, 'f', 7, 64)
	}
	if ix.Location.Lon > 0 {
		entry.Longitude = strconv.FormatFloat(ix.Location.Lon, 'f', 7, 64)
	}
	return entry
}

// flexDate creates an Entry Start/End value from a date object and precision indicator.
func flexDate(d time.Time, precision int) string {
	if d.Year() == 1 {
		return ""
	}
	s := d.Format("2006-01-02")
	if precision == 0 {
		s = s[:4]
	} else if precision == 1 {
		s = s[:7]
	}
	return s
}

// parseFlexDate converts an entry Start/End value into a date value and precision indicator.
func parseFlexDate(s string) (time.Time, int) {
	if s == "" {
		return time.Time{}, 0
	}
	precision := -1
	switch len(s) {
	case 4:
		precision = 0
	case 7:
		precision = 1
		s = s + "-01-01"
	case 10:
		precision = 2
		s = s + "-01"
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		//TODO: Log error
	}
	return t, precision
}

// Links returns a string slice of slugs that the entry identified by slug links to.
func (b *BleveSearch) Links(slug string) ([]string, error) {
	ret := []string{}
	doc, err := b.searchIndex.Document(slug)
	if err != nil || doc == nil {
		return ret, err
	}
	for _, field := range doc.Fields {
		switch field.Name() {
		case "Links":
			ret = append(ret, string(field.Value()))
		}
	}
	return ret, nil
}

// Stub returns indexed entry data for the given slug with truncated Description value and Links populated.
// GetEntryFromIndex returns an entry from the search index suitable for display.
func (b *BleveSearch) Stub(slug string) (model.Entry, error) {
	doc, err := b.searchIndex.Document(slug)
	if err != nil || doc == nil {
		return model.Entry{}, err
	}
	indexed := IndexedEntry{Custom: make(map[string]string)}
	for _, field := range doc.Fields {
		switch field.Name() {
		case "Name":
			indexed.Name = string(field.Value())
		case "Description":
			indexed.Description = string(field.Value())
		case "EntryType":
			indexed.EntryType = string(field.Value())
		case "Tags": // there's a separate Tags field for each tag value in a document
			indexed.Tags = append(indexed.Tags, string(field.Value()))
		case "LinksTo":
			indexed.Links = append(indexed.Links, string(field.Value()))
		case "Start":
			//TODO: Suggest an example of this at https://blevesearch.com/docs/Index-Mapping/
			df, ok := field.(*document.DateTimeField)
			if ok {
				dt, err := df.DateTime()
				if err == nil {
					indexed.Start = dt
				}
			}
		case "End":
			df, ok := field.(*document.DateTimeField)
			if ok {
				dt, err := df.DateTime()
				if err == nil {
					indexed.End = dt
				}
			}
		case "StartPrecision":
			indexed.StartPrecision, _ = strconv.Atoi(string(field.Value()))
		case "EndPrecision":
			indexed.EndPrecision, _ = strconv.Atoi(string(field.Value()))
		case "Address":
			indexed.Address = string(field.Value())
		case "Modified":
			df, ok := field.(*document.DateTimeField)
			if ok {
				dt, err := df.DateTime()
				if err == nil {
					indexed.Modified = dt
				}
			}
		default:
			if strings.HasPrefix(field.Name(), "Custom.") {
				key := strings.Split(field.Name(), ".")[1]
				indexed.Custom[key] = string(field.Value())
			}
		}
	}
	return indexed.Entry(), nil
}

// entryIndexMapping returns the default index settings for
// new and existing search indexes.
func (b *BleveSearch) entryIndexMapping() mapping.IndexMapping {
	entryMapping := bleve.NewDocumentMapping()
	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = en.AnalyzerName
	boolFieldMapping := bleve.NewBooleanFieldMapping()
	timeMapping := bleve.NewDateTimeFieldMapping()
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Type = "text"
	keywordFieldMapping.Analyzer = standard.Name
	precisionMapping := bleve.NewTextFieldMapping()
	precisionMapping.Type = "number"
	geoMapping := bleve.NewGeoPointFieldMapping()
	entryMapping.AddFieldMappingsAt("Name", englishTextFieldMapping)
	entryMapping.AddFieldMappingsAt("Description", englishTextFieldMapping)
	entryMapping.AddFieldMappingsAt("Tags", keywordFieldMapping)
	entryMapping.AddFieldMappingsAt("EntryType", keywordFieldMapping)
	entryMapping.AddFieldMappingsAt("Exclude", boolFieldMapping)
	entryMapping.AddFieldMappingsAt("Links", keywordFieldMapping)
	entryMapping.AddFieldMappingsAt("Start", timeMapping)
	entryMapping.AddFieldMappingsAt("End", timeMapping)
	entryMapping.AddFieldMappingsAt("StartPrecision", precisionMapping)
	entryMapping.AddFieldMappingsAt("EndPrecision", precisionMapping)
	entryMapping.AddFieldMappingsAt("Address", englishTextFieldMapping)
	entryMapping.AddFieldMappingsAt("Custom", englishTextFieldMapping)
	entryMapping.AddFieldMappingsAt("Modified", timeMapping)
	entryMapping.AddFieldMappingsAt("Location", geoMapping)
	//TODO: Index lat/long; create/mod date
	im := bleve.NewIndexMapping()
	im.AddDocumentMapping("Entry", entryMapping)
	return im
}

// initSearch should be called to setup search on application
// startup after entries are loaded/available.
func (b *BleveSearch) initSearch() error {
	indexPath := config.SearchPath() + "/index_meta.json"
	if localfs.PathExists(indexPath) {
		// open existing search index
		var err error
		b.searchIndex, err = bleve.Open(indexPath)
		if err != nil {
			return err
		}
	} else {
		if err := b.Rebuild(); err != nil {
			return err
		}
	}
	return nil
}

// IndexEntry adds or updates an entry in the index
func (b *BleveSearch) IndexEntry(entry model.Entry) error {
	indexed := NewIndexedEntry(entry)
	return b.searchIndex.Index(entry.Slug(), indexed)
}

// RemoveFromIndex removes an entry from the index
func (b *BleveSearch) RemoveFromIndex(slug string) error {
	return b.searchIndex.Delete(slug)
}

// Rebuild creates a new search index of current entries.
func (b *BleveSearch) Rebuild() error {
	if err := util.DelTree(config.SearchPath()); err != nil {
		return err
	}
	// create new search index
	var err error
	b.searchIndex, err = bleve.New(config.SearchPath(), b.entryIndexMapping())
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
		indexedEntry := NewIndexedEntry(entry)
		if err := b.searchIndex.Index(slug, indexedEntry); err != nil {
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

// ReverseLinks returns a list of slugs that link to the entry identified by `slug`.
func (b *BleveSearch) ReverseLinks(slug string) ([]string, error) {
	ret := []string{}
	matchQuery := bleve.NewMatchPhraseQuery(slug)
	matchQuery.SetField("Links")
	req := bleve.NewSearchRequestOptions(matchQuery, util.MaxInt32, 0, false)
	result, err := b.searchIndex.Search(req)
	if err != nil {
		return ret, err
	}
	hits := result.Hits
	for _, hit := range hits {
		ret = append(ret, hit.ID)
	}
	return ret, nil
}

// IndexedCount returns the total number of entries in the search index.
func (b *BleveSearch) IndexedCount() uint64 {
	i, _ := b.searchIndex.DocCount()
	return i
}

// SearchEntries returns a page of results based on multiple filters and search query.
func (b *BleveSearch) SearchEntries(types model.EntryTypes, keywords string, onlyTags []string,
	anyTags []string, sort SortOrder, pageNo int, pageSize int) (EntryResults, error) {
	q := b.buildSearchQuery(types, keywords, onlyTags, anyTags)
	req := bleve.NewSearchRequestOptions(q, pageSize, (pageNo-1)*pageSize, false)
	if sort == SortName {
		req.SortBy([]string{"Name"})
	} else if sort == SortRecent {
		req.SortBy([]string{"-Modified"})
	} else {
		req.SortBy([]string{"-_score"})
	}
	searchResult, err := b.searchIndex.Search(req)
	if err != nil {
		return EntryResults{}, err
	}
	ids := []string{}
	for _, hit := range searchResult.Hits {
		ids = append(ids, hit.ID)
	}
	results := EntryResults{Types: types, Search: keywords, AnyTags: anyTags, OnlyTags: onlyTags,
		Sort: sort, PageNo: pageNo, PageSize: pageSize, Total: searchResult.Total, Entries: []model.Entry{}}
	for _, id := range ids {
		entry, err := b.Stub(id)
		if err != nil {
			if _, notFound := err.(model.EntryNotFound); notFound {
				return EntryResults{}, errors.New("Document in search results not found in index: " + id)
			} else {
				return EntryResults{}, err
			}
		}
		results.Entries = append(results.Entries, entry)
	}
	return results, nil
}

// RefreshResults re-runs a search to freshen the results in case any entries have been modified.
func (b *BleveSearch) RefreshResults(stale EntryResults) (EntryResults, error) {
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
		entry, _ := b.Stub(hit.ID)
		ret = append(ret, entry)
	}
	return ret, nil
}

// BrokenLinks returns a map of all pages that link to non-existent pages. Each
// page with broken links is a key in the map, value is a string slice of slugs
// that don't match existing pages.
func (b *BleveSearch) BrokenLinks() (map[string][]string, error) {
	ret := make(map[string][]string)
	slugs, err := b.IndexedSlugs()
	if err != nil {
		return ret, err
	}
	for _, slug := range slugs {
		entryLinks, err := b.Links(slug)
		if err != nil {
			return ret, err
		}
		for _, link := range entryLinks {
			if !util.StringSliceContains(slugs, link) {
				if brokenLinks, exists := ret[slug]; exists {
					ret[slug] = append(brokenLinks, link)
				} else {
					ret[slug] = []string{link}
				}
			}
		}
	}
	return ret, nil
}
