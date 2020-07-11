/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/*
This file supports display of the CLI application.
*/

package display

import (
	"fmt"
	"math"
	"memory/app"
	"memory/app/model"
	"reflect"
	"strings"

	"github.com/buger/goterm"
	"github.com/mitchellh/go-wordwrap"
)

const prefix = "  "
const spacer = "  |  "

type page struct {
	startIndex int // index of first entry shown on the page
	count      int // number of entries shown on the page
}

// EntryPager is a stateful object to handle paging and display of an entry list
type EntryPager struct {
	results         app.EntryResults // entries to display and filter settings
	currentPage     int              // current page, 0-based
	pages           []page           // total number of pages
	renderedEntries [][]string       // rendered output for each entry
	header          []string         // rendered page header
	footer          []string         // rendered page footer
	screenHeight    int              // screen height at last render
	screenWidth     int              // screen width at last render
}

// NewEntryPager prepares a list of entries for paged display.
func NewEntryPager(results app.EntryResults) EntryPager {
	pager := EntryPager{results: results}
	updateRenderings(&pager)
	return pager
}

// PrintPage outputs the current page.
func (pager *EntryPager) PrintPage() {
	// re-render pages if the terminal size has changed
	if pager.screenHeight != goterm.Height() || pager.screenWidth != goterm.Width() {
		updateRenderings(pager)
		pager.currentPage = 0
	}
	fmt.Println(strings.Join(pager.header, "\n"))
	page := pager.pages[pager.currentPage]
	for i := page.startIndex; i <= page.startIndex+page.count; i++ {
		fmt.Println(strings.Join(pager.renderedEntries[i], "\n"))
	}
	fmt.Println(strings.Join(pager.footer, "\n"))
}

// updateRenderings creates arrays of output for header, footer and each entry
// so that paging can be established. This happens when a new struct is created
// or when PrintPage detects a change in window size.
func updateRenderings(pager *EntryPager) {
	pager.screenHeight = goterm.Height()
	pager.screenWidth = goterm.Width()
	pager.renderedEntries = renderEntries(*pager)
	// once to get a sense of lines needed to display header
	pager.header = renderHeader(*pager)
	pager.footer = renderFooter(*pager)
	pager.pages = calculatePages(*pager)
	// and again to include the calculated page count
	pager.header = renderHeader(*pager)
}

// addSettingToHeader is used by renderHeader to add a filter setting to the header. It returns
// the filter appended to the last line of the header or wraps to a new line if neeed.
func addSettingToHeader(pager EntryPager, header []string, label string, value string) []string {
	s := "|  " + label + ": " + value + "  "
	line := header[len(header)-1]
	if (len(line) + len(s) + 2) > displayWidth(pager) {
		// wrap to new line
		header = append(header, s[1:])
	} else {
		// append to last line
		header[len(header)-1] = line + s
	}
	return header
}

// renderHeader returns the top 2-3 rows of a page display.
// It should look something like this:
// --------------------------------------------------------------------------------------------------
//   14 results  |  Page 1 of 2  |  Showing: All types  |  Sort: Most recent  |  Containing: port
//
func renderHeader(pager EntryPager) []string {
	totalWidth := displayWidth(pager)
	// delcare return value and add top border
	lines := []string{strings.Repeat("-", totalWidth)}
	// info header template
	types := pager.results.Types.String()
	info := fmt.Sprintf("%4d results  |  Page %d of %d  |  Showing: %s  ",
		len(pager.results.Entries), pager.currentPage+1, len(pager.pages), types)
	lines = append(lines, info)
	// add sort
	if pager.results.Sort == app.SortName {
		lines = addSettingToHeader(pager, lines, "Sort", "Name")
	} else {
		lines = addSettingToHeader(pager, lines, "Sort", "Most recent")
	}
	// optional Tags filter
	if len(pager.results.Tags) > 0 {
		lines = addSettingToHeader(pager, lines, "Tagged with", strings.Join(pager.results.Tags, ", "))
	}
	// optional Contains filter
	if pager.results.Contains != "" {
		lines = addSettingToHeader(pager, lines, "Containing", pager.results.Contains)
	}
	// optional Starting With filter
	if pager.results.StartsWith != "" {
		lines = addSettingToHeader(pager, lines, "Starting with", pager.results.StartsWith)
	}
	// optional Search filter
	if pager.results.StartsWith != "" {
		lines = addSettingToHeader(pager, lines, "Search for", pager.results.Search)
	}
	// blank line at the bottom
	lines = append(lines, "")
	return lines
}

// renderFooter renders the footer that provides command options and should look
// something like this:
//
// Enter # to view details, [n]ext page, [p]revious page, [Q]uit
// >
func renderFooter(pager EntryPager) []string {
	lines := []string{"", "Enter # to view details"}
	if pager.currentPage < (len(pager.pages) - 1) {
		lines[0] = lines[0] + ", [n]ext page"
	}
	if pager.currentPage > 0 {
		lines[0] = lines[0] + ", [p]revious page"
	}
	lines[0] = lines[0] + ", [Q]uit"
	lines = append(lines, "> ")
	return lines
}

// displayWidth returns the total width of the display table.
func displayWidth(pager EntryPager) int {
	fw := float64(pager.screenWidth)
	return pager.screenWidth - int(math.Floor(fw*0.1))
}

// displayHeight returns the total height to be used.
func displayHeight(pager EntryPager) int {
	fh := float64(pager.screenHeight)
	return pager.screenHeight - int(math.Floor(fh*0.2))
}

// entryLines returns a string slice representing the lines of an individual entry listing.
// It should look something like this (lines vary by entry type and content):
//   1.  [Place] Rockport, MA
//       Tags: town, vacation
//       A seaside town on Cape Ann, North Shore of Massachusetts. We go there
//       every year for 4th of July and usually several other random times...
//       ----------------------------------------------------------------------
func renderEntry(pager EntryPager, ix int, entry model.Entry) []string {
	leftMargin := 6 // "  1.  "
	blankLeftMargin := strings.Repeat(" ", leftMargin)
	contentWidth := displayWidth(pager) - leftMargin
	// ex. Place
	typeName := strings.Title(reflect.TypeOf(entry).Name())
	// ex. "  1.  [Place] Rockport, MA"
	titleLine := fmt.Sprintf("%3d.  [%s] %s", ix+1, typeName, entry.Name())
	// `lines` will be the return value
	lines := []string{titleLine}
	// add Tags line, ex. "      Tags: town, vacation"
	if len(entry.Tags()) > 0 {
		tagLine := blankLeftMargin + "Tags: " + strings.Join(entry.Tags(), ", ")
		lines = append(lines, tagLine)
	}
	// add Description, ex. "      A seaside town..." - Max 2 lines w/ elipsis if truncated
	if entry.Description() != "" {
		descWrapped := wordwrap.WrapString(entry.Description(), uint(contentWidth))
		descLines := strings.Split(descWrapped, "\n")
		// add elipses to 2nd line if more than 2 lines and truncate array
		if len(descLines) > 2 {
			for len(descLines[1]) > (contentWidth - 3) {
				words := strings.Split(descLines[1], " ")
				words = words[:len(words)-1]
				descLines[1] = strings.Join(words, " ")
			}
			descLines[1] = descLines[1] + "..."
			descLines = descLines[:2]
		}
		for _, line := range descLines {
			lines = append(lines, blankLeftMargin+line)
		}
	}
	// add bottom border
	lines = append(lines, blankLeftMargin+strings.Repeat("-", contentWidth))
	return lines
}

// renderEntries calls renderEntry for each entry in results and returns an array of them
func renderEntries(pager EntryPager) [][]string {
	ret := make([][]string, len(pager.results.Entries))
	for i, entry := range pager.results.Entries {
		ret[i] = renderEntry(pager, i, entry)
	}
	return ret
}

// calculatePages returns a slice of page structs by figuring out how many entries
// can fit on each page given available screen height.
func calculatePages(pager EntryPager) []page {
	currentPage := page{startIndex: 0, count: 0}
	pages := []page{}
	headerFooterHeight := len(pager.header) + len(pager.footer)
	linesOnPage := headerFooterHeight
	for i, entryLines := range pager.renderedEntries {
		// start new page if we don't have space for this entry
		if (linesOnPage + len(entryLines)) > displayHeight(pager) {
			pages = append(pages, currentPage)
			currentPage = page{startIndex: i, count: 0}
			linesOnPage = headerFooterHeight
		}
		// add entry to current page
		currentPage.count = currentPage.count + 1
		linesOnPage = linesOnPage + len(entryLines)
	}
	pages = append(pages, currentPage)
	return pages
}
