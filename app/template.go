/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/* This file contains functions to translate Entry objects to and from
   string representations consisting of attributes in yaml frontmatter followed
   by the description. */

package app

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"
)

var tmpl *template.Template

// RenderEntryFile returns a string with attributes in yaml frontmatter followed by the description.
func RenderEntryFile(entry Entry) (string, error) {
	if tmpl == nil {
		var err error
		tmpl, err = template.New("Entry").Parse(Template)
		if err != nil {
			return "", errors.New("cannot compile template: " + err.Error())
		}
	}
	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, entry)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ParseEntryFile converts a string of yaml frontmatter followed by description into an Entry.
func ParseEntryFile(content string) (Entry, error) {
	// break the string into a slice of lines
	lines := strings.Split(content, "\n")
	// first line validation
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return Entry{}, errors.New("the first line of an entry must be ---")
	}
	// parse rest of file into temporary map
	attrs := make(map[string]string)
	for ix, line := range lines[1:] {
		// after metadata, everything else is description
		if strings.TrimSpace(line) == "---" {
			if len(lines) > ix+1 {
				attrs["description"] = strings.TrimSpace(strings.Join(lines[ix+2:], "\n"))
			} else {
				attrs["description"] = ""
			}
			break
		}
		// allow blank lines in metadata
		if strings.TrimSpace(line) == "" {
			continue
		}
		// validate meta data line
		if !strings.Contains(line, ":") {
			return Entry{}, errors.New("invalid attribute format (missing :) on line " + string(ix+1))
		}
		// parse the attribute and add it to the map
		attr := strings.SplitN(line, ":", 2)
		attrs[strings.ToLower(strings.TrimSpace(attr[0]))] = strings.TrimSpace(attr[1])
	}
	// initalize return value
	entry := Entry{}
	// validate and set attributes
	// Description
	if desc, exists := attrs["description"]; exists {
		entry.Description = desc
	} else {
		return Entry{}, errors.New("attributes must be separated from decsription with a --- line")
	}
	// Type
	if t, exists := attrs["type"]; !exists {
		return Entry{}, errors.New("missing required 'type' attribute")
	} else if t != EntryTypeEvent && t != EntryTypePerson && t != EntryTypePlace &&
		t != EntryTypeThing && t != EntryTypeNote {
		return Entry{}, fmt.Errorf("'type' is not one of the valid entry types: %s, %s, %s, %s, %s",
			EntryTypeEvent, EntryTypePerson, EntryTypePlace, EntryTypeThing, EntryTypeNote)
	} else {
		entry.Type = t
	}
	// Name
	if name, exists := attrs["name"]; !exists {
		return Entry{}, errors.New("missing required 'name' attribute")
	} else {
		entry.Name = name
	}
	// Tags
	if tags, exists := attrs["tags"]; exists {
		// trim of brackets and split on comma
		entry.Tags = processTags(tags)
	}
	return entry, nil
}

// Template is a generic entry template.
var Template = `---
Name: {{.Name}}
{{if eq .Type "Event"}}start: 
{{end}}Type: {{.Type}}
Tags: {{.TagsString}}
---

{{.Description}}
`

// processTags takes in a comma-separated string and returns a slice of trimmed values
func processTags(tags string) []string {
	if strings.HasPrefix(tags, "[") && strings.HasPrefix(tags, "]") {
		tags = tags[1 : len(tags)-1]
	}
	if strings.TrimSpace(tags) == "" {
		return []string{}
	}
	arr := strings.Split(tags, ",")
	for i, tag := range arr {
		arr[i] = strings.ToLower(strings.TrimSpace(tag))
	}
	return arr
}
