/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/* This file contains functions to translate Entry objects to and from
   string representations consisting of attributes in yaml frontmatter followed
   by the description. */

package template

import (
	"bytes"
	"errors"
	"fmt"
	"memory/app/model"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

var tmpl *template.Template

// Template is a generic entry template.
var Template = `---
Name: {{.Name}}
Type: {{.Type}}
Tags: {{.TagsString}}
{{if eq .Type "Event"}}Start: {{.Start}}
End: {{.End}}
{{end}}{{if eq .Type "Place"}}Address: {{.Address}}
Latitude: {{.Latitude}}
Longitude: {{.Longitude}}
{{end}}{{range $key, $val := .Custom}}{{$key}}: {{$val}}
{{end}}---

{{.Description}}
`

// RenderYamlDown returns a string with attributes in yaml frontmatter followed by the description.
func RenderYamlDown(entry model.Entry) (string, error) {
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

// ParseYamlDown converts a string of yaml frontmatter followed by description into an Entry.
func ParseYamlDown(content string) (model.Entry, error) {
	// break the string into a slice of lines
	lines := strings.Split(content, "\n")
	// first line validation
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return model.Entry{}, errors.New("the first line of an entry must be ---")
	}
	// parse rest of file into temporary map
	attrs := make(map[string]string)
	for ix, line := range lines[1:] {
		// after metadata, everything else is description
		if strings.TrimSpace(line) == "---" {
			if len(lines) > ix+1 {
				attrs["_description"] = strings.TrimSpace(strings.Join(lines[ix+2:], "\n"))
			} else {
				attrs["_description"] = ""
			}
			break
		}
		// allow blank lines in metadata
		if strings.TrimSpace(line) == "" {
			continue
		}
		// validate meta data line
		if !strings.Contains(line, ":") {
			return model.Entry{}, errors.New("invalid attribute format (missing :) on line " + fmt.Sprint(ix+1))
		}
		// parse the attribute and add it to the map
		attr := strings.SplitN(line, ":", 2)
		attrs[strings.TrimSpace(attr[0])] = strings.TrimSpace(attr[1])
	}
	// initalize return value
	entry := model.Entry{}
	// validate Description
	if val, exists := attrs["_description"]; exists {
		entry.Description = val
	} else {
		return model.Entry{}, errors.New("attributes must be separated from decsription with a --- line")
	}
	// validate Name
	if name, exists := attrs["Name"]; exists {
		if err := model.ValidateEntryName(name); err != nil {
			return model.Entry{}, err
		}
		entry.Name = name
	} else {
		return model.Entry{}, errors.New("missing required Name attribute")
	}
	// validate Type
	if t, exists := attrs["Type"]; !exists {
		return model.Entry{}, errors.New("missing required Type attribute")
	} else if t != model.EntryTypeEvent && t != model.EntryTypePerson && t != model.EntryTypePlace &&
		t != model.EntryTypeThing && t != model.EntryTypeNote {
		return model.Entry{}, fmt.Errorf("Type is not one of the valid entry types (%s, %s, %s, %s, %s)",
			model.EntryTypeEvent, model.EntryTypePerson, model.EntryTypePlace, model.EntryTypeThing, model.EntryTypeNote)
	} else {
		entry.Type = t
	}
	// handle optional attributes
	for key, val := range attrs {
		switch key {
		case "Name", "Type", "_description":
			// handled above
		case "Tags":
			// trim of brackets and split on comma
			entry.Tags = processTags(val)
		case "Start", "End":
			matched, err := regexp.Match(`([\d]{4})?(-[\d]{2})?(-[\d]{2})?`, []byte(val))
			if err != nil || !matched {
				return model.Entry{}, errors.New("value for " + key + " is invalid: must be YYYY, YYYY-MM or YYYY-MM-DD")
			}
			if key == "Start" && val == "" {
				return model.Entry{}, errors.New("value is required for " + key)
			}
			if key == "Start" {
				entry.Start = val
			} else {
				entry.End = val
			}
		case "Latitude", "Longitude":
			if val != "" {
				if _, err := strconv.ParseFloat(val, 64); err != nil {
					return model.Entry{}, errors.New("value for " + key + " is invalid")
				}
				if key == "Latitude" {
					entry.Latitude = val
				} else {
					entry.Longitude = val
				}
			}
		case "Address":
			entry.Address = val
		default:
			if entry.Custom == nil {
				entry.Custom = make(map[string]string)
			}
			entry.Custom[key] = val
		}
	}
	return entry, nil
}

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
		arr[i] = strings.TrimSpace(tag)
	}
	return arr
}
