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

package link

import (
	"fmt"
	"memory/app"
	"memory/app/model"
	"memory/app/util"
	"regexp"
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
			name = strings.ReplaceAll(name, "  ", " ")
		}
		// remove ! if it's already there (! prefix indicates non-existent entry)
		hadBang := false
		if strings.HasPrefix(name, "!") {
			name = name[1:]
			hadBang = true
		}
		// add to results if exists, otherwise add ! prefix
		if _, exists := app.GetEntry(name); exists {
			if !util.StringSliceContains(links, name) {
				links = append(links, name)
			}
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
	}
	return parsed, links
}

// ResolveLinks accepts a slice of Entry names and returns
// a slice of Entries that exist with those names.
func ResolveLinks(links []string) []model.Entry {
	resolved := []model.Entry{}
	for _, name := range links {
		if entry, exists := app.GetEntry(name); exists {
			resolved = append(resolved, entry)
		}
	}
	return resolved
}

// PopulateLinks populates the LinksTo and LinkedFrom slices on all entries by
// parsing the descriptions for links.
func PopulateLinks() {
	//TODO implement PopulateLinks
}

// LinkedFrom returns a slice of entry names that link to
// the given entry name. Assumes all entries have already
// been parsed.
func LinkedFrom(name string) []string {
	//TODO implement LinkedFrom
	return nil
}
