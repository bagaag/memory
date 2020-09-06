/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/* Linker interface manages parsing and lookup of links between entries. */

package links

import (
	"memory/util"
	"regexp"
	"strings"
)

var linkExp *regexp.Regexp // initialized on first use

// LinkRegExp returns the Regexp used to find links within entry descriptions.
func LinkRegExp() (*regexp.Regexp, error) {
	if linkExp == nil {
		var err error
		linkExp, err = regexp.Compile("\\[([[:alnum:]?][^~\\]]*)\\]\\(?")
		if err != nil {
			return nil, err
		}
	}
	return linkExp, nil
}

// RenderLinks parses the links in a string and returns the string with
// updated links rendered to indicate existence or non-existence of the links
// based on the result of the exists function.
func RenderLinks(s string, exists func(string) bool) string {
	// init return values
	parsed := s
	linkExp, _ := LinkRegExp()
	// get [links]
	results := linkExp.FindAllStringIndex(s, -1)
	for _, pair := range results {
		link := s[pair[0]:pair[1]]
		// ignore external links, which are followed immediately by "("
		if strings.HasSuffix(link, "(") {
			continue
		}
		// strip off brackets, remove line breaks and consecutive spaces
		name := link[1 : len(link)-1]
		name = strings.ReplaceAll(name, "\n", " ")
		for strings.Contains(name, "  ") {
			name = strings.ReplaceAll(name, "  ", " ")
		}
		// remove ? if it's already there (? prefix indicates non-existent entry)
		hadBang := false
		if strings.HasPrefix(name, "?") {
			name = name[1:]
			hadBang = true
		}
		slug := util.GetSlug(name)
		// add to results if exists, otherwise add ! prefix
		if exists(slug) {
			// remove erroneous ? prefix if needed
			if hadBang {
				linkWithoutBang := "[" + link[2:]
				parsed = strings.Replace(parsed, link, linkWithoutBang, 1)
			}
		} else if !hadBang {
			// entry doesn't exist, add a ? if needed
			link404 := "[?" + link[1:]
			parsed = strings.Replace(parsed, link, link404, 1)
		}
	}
	return parsed
}

// ParseLinks looks for [Name] links within the given string and
// returns a slice of index pairs found. Links that cannot be
// resolved are replaced with a ! prefix in the parsed return
// value, as in [!Not Found].
func ExtractLinks(s string) []string {
	// init return values
	list := []string{}
	linkExp, err := LinkRegExp()
	if err != nil {
		//TODO: Log error
		return []string{}
	}
	// get [links]
	results := linkExp.FindAllStringIndex(s, -1)
	for _, pair := range results {
		link := s[pair[0]:pair[1]]
		// ignore external links, which are followed immediately by "("
		if strings.HasSuffix(link, "(") {
			continue
		}
		// strip off brackets, remove line breaks and consecutive spaces
		name := link[1 : len(link)-1]
		name = strings.ReplaceAll(name, "\n", " ")
		for strings.Contains(name, "  ") {
			name = strings.ReplaceAll(name, "  ", " ")
		}
		if strings.HasPrefix(name, "?") {
			name = name[1:]
		}
		slug := util.GetSlug(name)
		if !util.StringSliceContains(list, name) {
			list = append(list, slug)
		}
	}
	return list
}
