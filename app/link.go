/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/*
Contains logic to parse, store and update entry
links within description fields using the form [Entry Name].
Brackets can be used for non-linking purposes in Description
fields by prefixing a ~ as in [~Not a link], wich will be
displayed without the tilde. Links that cannot be resolved
will be replaced with a ! prefix as in [!Entry Name].
*/

package app

import (
	"fmt"
	"memory/util"
	"regexp"
	"sort"
	"strings"
)

var linkExp *regexp.Regexp

// ParseLinks looks for [Name] links within the given string and
// returns a slice of index pairs found. Links that cannot be
// resolved are replaced with a ! prefix in the parsed return
// value, as in [!Not Found].
func ParseLinks(s string) (string, []string) {
	// init return values
	parsed := s
	links := []string{}
	// compile links regexp
	if linkExp == nil {
		var err error
		linkExp, err = regexp.Compile("\\[([[:alnum:]!][^~\\]]*)\\]")
		if err != nil {
			fmt.Println("Error compiling link regexp:", err)
			return s, []string{}
		}
	}
	// get [links]
	results := linkExp.FindAllStringIndex(s, -1)
	for _, pair := range results {
		link := s[pair[0]:pair[1]]
		// strip off brackets, remove line breaks and consecutive spaces
		name := link[1 : len(link)-1]
		name = strings.ReplaceAll(name, "\n", " ")
		for strings.Contains(name, "  ") {
			name = strings.ReplaceAll(name, "  ", " ") //TODO: use regex to replace 2+ whitespace
		}
		// remove ! if it's already there (! prefix indicates non-existent entry)
		hadBang := false
		if strings.HasPrefix(name, "!") {
			name = name[1:]
			hadBang = true
		}
		// add to results if exists, otherwise add ! prefix
		if _, exists := GetEntry(name); exists {
			// remove erroneous ! prefix if needed
			if hadBang {
				linkWithoutBang := "[" + link[2:]
				parsed = strings.Replace(parsed, link, linkWithoutBang, 1)
			}
		} else if !hadBang {
			// entry doesn't exist, add a ! if needed
			link404 := "[!" + link[1:]
			parsed = strings.Replace(parsed, link, link404, 1)
		}
		if !util.StringSliceContains(links, name) {
			links = append(links, name)
		}
	}
	return parsed, links
}

// ResolveLinks accepts a slice of Entry names and returns
// a slice of Entries that exist with those names.
func ResolveLinks(links []string) []Entry {
	resolved := []Entry{}
	for _, name := range links {
		if entry, exists := GetEntry(name); exists {
			resolved = append(resolved, entry)
		}
	}
	return resolved
}

// populateLinks populates the LinksTo and LinkedFrom slices on all entries by
// parsing the descriptions for links. Assumes the calling function already
// has a lock on data.
func populateLinks() {
	fromLinks := make(map[string][]string)
	for _, entry := range data.Names {
		// parse and save outgoing links for this entry
		searchText := entry.Description
		newDesc, links := ParseLinks(searchText)
		entry.Description = newDesc
		entry.LinksTo = links
		data.Names[GetSlug(entry.Name)] = entry
		// add links in reverse direction
		fromName := entry.Name
		for _, toName := range links {
			names, exists := fromLinks[GetSlug(toName)]
			if !exists {
				names = []string{fromName}
			} else if !util.StringSliceContains(names, fromName) {
				names = append(names, fromName)
			}
			fromLinks[toName] = names
		}
	}
	// save the fromLinks in corresponding entries
	for name, linkedFrom := range fromLinks {
		entry, exists := GetEntry(name)
		if exists {
			entry.LinkedFrom = linkedFrom
			data.Names[GetSlug(entry.Name)] = entry
		}
	}
}

// BrokenLinks returns a map of string slices containing names of linked-to pages that don't
// exist; the name of the page containing the link is the key.
func BrokenLinks() map[string][]string {
	ret := make(map[string][]string)
	data.lock()
	defer data.unlock()
	for fromName, fromEntry := range data.Names {
		for _, toName := range fromEntry.LinksTo {
			if _, entryExists := data.Names[GetSlug(toName)]; !entryExists {
				var brokenLinks []string
				var existingList bool
				if brokenLinks, existingList = ret[fromName]; existingList {
					brokenLinks = append(brokenLinks, toName)
					sort.Strings(brokenLinks)
				} else {
					brokenLinks = []string{toName}
				}
				ret[fromName] = brokenLinks
			}
		}
	}
	return ret
}
