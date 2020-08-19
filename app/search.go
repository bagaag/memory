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
	"fmt"
	"memory/app/config"
	"memory/app/persist"

	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"

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
	englishTextFieldMapping.Analyzer = en.AnalyzerName
	entryMapping.AddFieldMappingsAt("Name", englishTextFieldMapping)
	entryMapping.AddFieldMappingsAt("Description", englishTextFieldMapping)
	entryMapping.AddFieldMappingsAt("Tags", englishTextFieldMapping)
	//TODO: Index event dates, modification date, lat/long, description excerpt
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
	query := bleve.NewMatchQuery(keywords)
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
	for key, entry := range data.Names {
		if err := searchIndex.Index(key, entry); err != nil {
			fmt.Println("Error indexing:", err)
		} else {
			count = count + 1
		}
	}
	fmt.Printf("Indexed %d out of %d entries.\n", count, len(data.Names))
	return nil
}
