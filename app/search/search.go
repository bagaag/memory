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

package search

import (
	"memory/app/model"
)

type Searcher interface {
	BrokenLinks() (map[string][]string, error)
	IndexEntry(entry model.Entry) error
	IndexedCount() uint64
	IndexedSlugs(prefix string) ([]string, error)
	IndexedNames(prefix string) ([]string, error)
	Links(slug string) ([]string, error)
	Rebuild() error
	RefreshResults(stale EntryResults) (EntryResults, error)
	RemoveFromIndex(slug string) error
	ReverseLinks(string) ([]string, error)
	SearchEntries(types model.EntryTypes, search string, onlyTags []string, anyTags []string,
		sort SortOrder, pageNo int, pageSize int) (EntryResults, error)
	Stub(slug string) (model.Entry, error)
	Timeline(start string, end string) ([]model.Entry, error)
}

// EntryResults is used to contain the results of GetEntries and the settings used
// to generate those results.
type EntryResults struct {
	Entries  []model.Entry
	Types    model.EntryTypes
	Search   string
	AnyTags  []string
	OnlyTags []string
	Sort     SortOrder
	Total    uint64
	PageNo   int
	PageSize int
}

// SortOrder is used to indicate one of the Sort constants
type SortOrder int

// SortScore sorts entries by search score
const SortScore = SortOrder(0)

// SortRecent sorts entries by descending modified date
const SortRecent = SortOrder(1)

// SortName sorts entries alphabetically by name
const SortName = SortOrder(2)
