/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/*
This file supports display of the CLI application.
*/

package cmd

import (
	"fmt"
	"math"
	"memory/app/model"
	"memory/app/search"
	"memory/util"
	"os"
	"strings"

	"github.com/buger/goterm"
	"github.com/mitchellh/go-wordwrap"
	"github.com/olekukonko/tablewriter"
)

const prefix = "  "
const spacer = "  |  "
const linesPerEntry = 5

// Page is described as the index of the first element displayed on the page and
// the number of elements displayed on the page.
type Page struct {
	StartIndex int // index of first entry shown on the page
	Count      int // number of entries shown on the page
}

// EntryPager is a stateful object to handle paging and display of an entry list
type EntryPager struct {
	Results         search.EntryResults // entries to display and filter settings
	pageCount       int                 // total number of pages
	renderedEntries [][]string          // rendered output for each entry on the current page
	header          []string            // rendered page header
	footer          []string            // rendered page footer
	screenHeight    int                 // screen height at last render
	screenWidth     int                 // screen width at last render
}

// NewEntryPager prepares a list of entries for paged display.
func NewEntryPager(results search.EntryResults) EntryPager {
	pager := EntryPager{Results: results}
	updateRenderings(&pager)
	return pager
}

// PrintPage outputs the current page.
func (pager *EntryPager) PrintPage() {
	// re-render pages if the has changed
	if pager.screenHeight != goterm.Height() || pager.screenWidth != goterm.Width() {
		setPageNumber(pager, 1)
		updateRenderings(pager)
	}
	fmt.Println(strings.Join(pager.header, "\n"))
	if len(pager.Results.Entries) == 0 {
		return
	}
	for i, entry := range pager.Results.Entries {
		lines := renderEntry(pager, i, entry)
		for _, line := range lines {
			fmt.Println(line)
		}
	}
	fmt.Println(strings.Join(pager.footer, "\n"))
}

// Next returns false if we're on the last page, otherwise
// true and advances to the next page.
func (pager *EntryPager) Next() bool {
	// see if there are enough entries to expand into another page
	if pager.Results.Total <= uint64(pager.Results.PageNo*pager.Results.PageSize) {
		return false
	}
	// increment page number and refresh results
	if !setPageNumber(pager, pager.Results.PageNo+1) {
		return false
	}
	pager.header = renderHeader(pager)
	pager.footer = renderFooter(pager)
	return true
}

// setPageNumber changes the current page to the specified number
// and returns Boolean indicating success
func setPageNumber(pager *EntryPager, pageNo int) bool {
	pager.Results.PageNo = pageNo
	var err error
	pager.Results, err = memApp.Search.RefreshResults(pager.Results)
	if err != nil {
		fmt.Printf("ERROR at setPageNumber(%d): %s", pageNo, err)
		return false
	}
	return true
}

// Prev returns true if we're on the first page, otherwise
// true and goes to the previous page.
func (pager *EntryPager) Prev() bool {
	if pager.Results.PageNo == 1 {
		return false
	}
	// increment page number and refresh results
	if !setPageNumber(pager, pager.Results.PageNo-1) {
		return false
	}
	pager.header = renderHeader(pager)
	pager.footer = renderFooter(pager)
	return true
}

// updateRenderings creates arrays of output for header, footer and each entry
// so that paging can be established. This happens when a new struct is created
// or when PrintPage detects a change in window size.
func updateRenderings(pager *EntryPager) {
	pager.screenHeight = goterm.Height()
	pager.screenWidth = goterm.Width()
	pager.pageCount = int(math.Ceil(float64(pager.Results.Total) / float64(pager.Results.PageSize)))
	pager.header = renderHeader(pager)
	pager.footer = renderFooter(pager)
}

// addSettingToHeader is used by renderHeader to add a filter setting to the header. It returns
// the filter appended to the last line of the header or wraps to a new line if neeed.
func addSettingToHeader(pager *EntryPager, header []string, label string, value string) []string {
	s := "|  " + label + ": " + value + "  "
	line := header[len(header)-1]
	if (len(line) + len(s) + 2) > displayWidth() {
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
func renderHeader(pager *EntryPager) []string {
	totalWidth := displayWidth()
	// delcare return value and add top border
	lines := []string{strings.Repeat("-", totalWidth)}
	// info header template
	types := pager.Results.Types.String()
	info := fmt.Sprintf("%4d results  |  Page %d of %d  |  Showing: %s  ",
		pager.Results.Total, pager.Results.PageNo, pager.pageCount, types)
	lines = append(lines, info)
	// add sort
	if pager.Results.Sort == search.SortName {
		lines = addSettingToHeader(pager, lines, "Sort", "Name")
	} else if pager.Results.Sort == search.SortRecent {
		lines = addSettingToHeader(pager, lines, "Sort", "Most recent")
	} else {
		lines = addSettingToHeader(pager, lines, "Sort", "Score")
	}
	// optional Any Tags filter
	if len(pager.Results.AnyTags) > 0 {
		lines = addSettingToHeader(pager, lines, "Any tags", strings.Join(pager.Results.AnyTags, ", "))
	}
	// optional Only Tags filter
	if len(pager.Results.OnlyTags) > 0 {
		lines = addSettingToHeader(pager, lines, "Only tags", strings.Join(pager.Results.OnlyTags, ", "))
	}
	// optional Search filter
	if pager.Results.Search != "" {
		lines = addSettingToHeader(pager, lines, "Search for", pager.Results.Search)
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
func renderFooter(pager *EntryPager) []string {
	lines := []string{""}
	cmd := "Enter # to view details"
	if pager.Results.PageNo < pager.pageCount {
		cmd = cmd + ", [n]ext page"
	}
	if pager.Results.PageNo > 1 {
		cmd = cmd + ", [p]revious page"
	}
	cmd = cmd + ", [Q]uit"
	lines = append(lines, cmd)
	return lines
}

// displayWidth returns the total width of the display table.
func displayWidth() int {
	fw := float64(goterm.Width())
	return int(fw - math.Floor(fw*0.1))
}

// displayHeight returns the total height to be used.
func displayHeight() int {
	fh := float64(goterm.Height())
	return int(fh - math.Floor(fh*0.1))
}

// entryLines returns a string slice representing the lines of an individual entry listing.
// It should look something like this (lines vary by entry type and content):
//   1.  [Place] Rockport, MA
//       Tags: town, vacation
//       A seaside town on Cape Ann, North Shore of Massachusetts. We go there
//       every year for 4th of July and usually several other random times...
//       ----------------------------------------------------------------------
func renderEntry(pager *EntryPager, ix int, entry model.Entry) []string {
	ix = ix + 1
	if ix == 10 {
		ix = 0
	}
	leftMargin := 6 // "  1.  "
	blankLeftMargin := strings.Repeat(" ", leftMargin)
	contentWidth := displayWidth() - leftMargin
	// ex. "  1.  [Place] Rockport, MA"
	titleLine := fmt.Sprintf("%3d.  [%s] %s", ix, entry.Type, entry.Name)
	// `lines` will be the return value
	lines := []string{titleLine}
	// add Tags line, ex. "      Tags: town, vacation"
	if len(entry.Tags) > 0 {
		tagLine := blankLeftMargin + "Tags: " + strings.Join(entry.Tags, ", ")
		lines = append(lines, tagLine)
	}
	// add Description, ex. "      A seaside town..." - Max 2 lines w/ elipsis if truncated
	if entry.Description != "" {
		descWrapped := wordwrap.WrapString(entry.Description, uint(contentWidth))
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

// entriesPerPage returns the number of ls entry results that can fit on each page.
func entriesPerPage(pager *EntryPager) int {
	headerFooterHeight := len(pager.header) + len(pager.footer)
	leftover := displayHeight() - headerFooterHeight
	result := math.Floor(float64(leftover / linesPerEntry))
	return int(math.Min(result, 10))
}

// ListPageSize estimates number of ls results that will fit within the current screen size
// before we have a populated pager to give us exact values for header/footer height.
func ListPageSize() int {
	headerHeight := 3
	footerHeight := 3
	leftover := displayHeight() - headerHeight - footerHeight
	result := math.Floor(float64(leftover / linesPerEntry))
	return int(math.Min(result, 10))
}

// EntryTables displays a table of entries, used when we're dumping all results after
// a non-interactive ls request, or when displaying a single entry details.
func EntryTables(entries []model.Entry) {
	width := goterm.Width() - 30
	fmt.Println("") // prefix with blank line
	for ix, entry := range entries {
		// holds table contents
		data := [][]string{}
		// add note name and type rows
		data = append(data, []string{"Name", entry.Name})
		data = append(data, []string{"Type", entry.Type})
		// tags row
		if len(entry.Tags) > 0 {
			data = append(data, []string{"Tags", strings.Join(entry.Tags, ", ")})
		}
		if entry.Start != "" {
			data = append(data, []string{"Start", entry.Start})
		}
		if entry.End != "" {
			data = append(data, []string{"End", entry.End})
		}
		if entry.Address != "" {
			data = append(data, []string{"Address", entry.Address})
		}
		if entry.Latitude != "" {
			data = append(data, []string{"Latitude", entry.Latitude})
		}
		if entry.Longitude != "" {
			data = append(data, []string{"Longitude", entry.Longitude})
		}
		// create and configure table
		table := tablewriter.NewWriter(os.Stdout)
		// add border to top unless this is the first
		if ix == len(entries)-1 {
			table.SetBorders(tablewriter.Border{Left: false, Top: true, Right: false, Bottom: true})
		} else {
			table.SetBorders(tablewriter.Border{Left: false, Top: true, Right: false, Bottom: false})
		}
		table.SetRowLine(false)
		table.SetColMinWidth(0, 8)
		table.SetColMinWidth(1, width)
		table.SetColWidth(width)
		table.SetAutoWrapText(true)
		table.SetReflowDuringAutoWrap(true)
		// add data and render
		table.AppendBulk(data)
		table.Render()
		fmt.Println(util.Indent(entry.Description, 2))
	}
	fmt.Println("") // finish with blank line
}

// EntryTable displays a single entry with full detail
func EntryTable(entry model.Entry) {
	entries := []model.Entry{entry}
	EntryTables(entries)
}

// LinksMenu displays a list of entry names in its LinksTo
// and LinkedFrom slices along with numbers for selection.
func LinksMenu(entry model.Entry) {
	fmt.Printf("\nLinks for %s [%s]\n\n", entry.Name, entry.Type)
	ix := 1
	if len(entry.LinksTo) > 0 {
		fmt.Println("  Links to:")
		for _, name := range entry.LinksTo {
			entry, _ := memApp.GetEntry(util.GetSlug(name))
			if entry.Type == "" {
				entry.Type = "?"
			}
			fmt.Printf("    %2d. %s [%s]\n", ix, name, entry.Type)
			ix = ix + 1
		}
		fmt.Println("")
	}
	if len(entry.LinkedFrom) > 0 {
		fmt.Println("  Linked from:")
		for _, name := range entry.LinkedFrom {
			entry, _ := memApp.GetEntry(name)
			if entry.Type == "" {
				entry.Type = "?"
			}
			fmt.Printf("    %2d. %s [%s]\n", ix, name, entry.Type)
			ix = ix + 1
		}
		fmt.Println("")
	}
}

// MissingLinkMenu presents a list of entry types that can be created for
// a non-existant entry name.
func MissingLinkMenu(name string) {
	fmt.Printf("\nEntry named '%s' does not exist.\n", name)
	fmt.Println("  1. Event")
	fmt.Println("  2. Person")
	fmt.Println("  3. Place")
	fmt.Println("  4. Thing")
	fmt.Println("  5. Note")
	fmt.Println("")
	fmt.Println("Enter 1-5 to create a new entry with this name, [b]ack or [Q]uit")
}

// WelcomeMessage personalizes the app with a message tailored to the visitors current journey.
//TODO: Flesh out the welcome journey
func WelcomeMessage() {
	fmt.Printf("Welcome. You have %d entries under management. "+
		"Type 'help' for assistance.\n", memApp.Search.IndexedCount())
}
