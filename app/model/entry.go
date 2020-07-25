/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package model

import "time"

// Entry is the base for all other memory content models (people, places, things, events)
type Entry struct {
	name        string
	description string
	tags        []string
	linksTo     []string
	linkedFrom  []string
	created     time.Time
	modified    time.Time
}

// IEntry is the interface that provides generic access to people, places, things, events
type IEntry interface {
	Name() string
	SetName(name string)
	Description() string
	SetDescription(description string)
	Tags() []string
	SetTags(tags []string)
	LinksTo() []string
	SetLinksTo(names []string)
	LinkedFrom() []string
	SetLinkedFrom(names []string)
	Created() time.Time
	Modified() time.Time
}

// NewEntry creates and returns a new Entry with the provided values.
func NewEntry(name string, description string, tags []string) Entry {
	now := time.Now()
	entry := Entry{
		name:        name,
		description: description,
		tags:        tags,
		created:     now,
		modified:    now,
	}
	return entry
}

// Name getter.
func (entry Entry) Name() string {
	return entry.name
}

// SetName setter.
func (entry *Entry) SetName(name string) {
	entry.name = name
	entry.modified = time.Now()
}

// Description getter.
func (entry Entry) Description() string {
	return entry.description
}

// SetDescription setter.
func (entry *Entry) SetDescription(description string) {
	entry.description = description
	entry.modified = time.Now()
}

// Tags getter.
func (entry Entry) Tags() []string {
	return entry.tags
}

// SetTags setter.
func (entry *Entry) SetTags(tags []string) {
	entry.tags = tags
	entry.modified = time.Now()
}

// LinksTo getter. These are entry names this entry links to.
func (entry Entry) LinksTo() []string {
	return entry.linksTo
}

// SetLinksTo setter.
func (entry *Entry) SetLinksTo(names []string) {
	entry.linksTo = names
	entry.modified = time.Now()
}

// LinkedFrom getter. These are entry names that link to this entry.
func (entry Entry) LinkedFrom() []string {
	return entry.linkedFrom
}

// SetLinkedFrom setter.
func (entry *Entry) SetLinkedFrom(names []string) {
	entry.linkedFrom = names
	entry.modified = time.Now()
}

// Created getter.
func (entry Entry) Created() time.Time {
	return entry.created
}

// Modified getter.
func (entry Entry) Modified() time.Time {
	return entry.modified
}
