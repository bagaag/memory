/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/* This file contains functions related to tag management. */

package app

import (
	"memory/util"
	"sort"
)

// GetTags returns a map of all defined tags, each with a sorted slice of
// associated entry names.
func GetTags() map[string][]string {
	tags := make(map[string][]string)
	for _, entry := range data.Names {
		for _, tag := range entry.Tags {
			names, exists := tags[tag]
			if !exists {
				names = []string{entry.Name}
			} else {
				if !util.StringSliceContains(names, entry.Name) {
					names = append(names, entry.Name)
					sort.Strings(names)
				}
			}
			tags[tag] = names
		}
	}
	return tags
}

// GetSortedTags takes the output of GetTags and returns a sorted
// slice of tags.
func GetSortedTags(tags map[string][]string) []string {
	keys := []string{}
	for tag := range tags {
		keys = append(keys, tag)
	}
	sort.Strings(keys)
	return keys
}
