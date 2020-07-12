/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/*
This file contains variables and functions used by the
`ls` command in both command line and interactive modes.
*/

package cmd

import (
	"fmt"
	"memory/app"
	"memory/app/config"
	"memory/cmd/display"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// flag values
var flagLsStartsWith string
var flagLsContains string
var flagLsLimit int
var flagLsSortModifiedDesc bool
var flagLsSortName bool
var flagLsFull bool
var flagLsTags []string
var flagLsTypes []string

//var flagLsTypes []string

// resetFlags returns all flag values to their defaults after being set via
// an interactive command (see lsInteractive).
func resetLsFlags() {
	flagLsStartsWith = ""
	flagLsContains = ""
	flagLsLimit = 0
	flagLsSortModifiedDesc = false
	flagLsSortName = false
	flagLsFull = false
	flagLsTags = []string{}
	flagLsTypes = []string{}
}

// sortOrder translates the various bool sort flags into a SortOrder value
func sortOrder() app.SortOrder {
	sort := app.SortRecent
	if flagLsSortName {
		sort = app.SortName
	}
	return sort
}

// lsCmd lists notes
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Displays and lists entries",
	Long:  `By default, lists 10 most recent entries of any type. Use flags to modify the listing.`,
	Run: func(cmd *cobra.Command, args []string) {
		types := parseTypes()
		results := app.GetEntries(types, flagLsStartsWith, flagLsContains, "", []string{}, sortOrder(), flagLsLimit)
		display.EntryTables(results.Entries)
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
	lsCmd.Flags().StringSliceVarP(&flagLsTypes, "types", "t", []string{},
		"Limit entries to one or more types (event, person, place, thing, note)")
	lsCmd.Flags().StringVar(&flagLsStartsWith, "starts-with", "",
		"Filter output with case-insensitive prefix")
	lsCmd.Flags().StringVarP(&flagLsContains, "contains", "c", "",
		"Filter output with case-insensitive substring")
	lsCmd.Flags().StringSliceVarP(&flagLsTags, "tags", "g", []string{},
		"Limit entries to those tagged with any of these")
	lsCmd.Flags().IntVarP(&flagLsLimit, "limit", "l", 0,
		"Specify maximum number of results to return")
	lsCmd.Flags().BoolVar(&flagLsSortModifiedDesc, "sort-modified-desc", false,
		"Sort results with most recently modified entries at the top (default)")
	lsCmd.Flags().BoolVar(&flagLsSortName, "sort-name", false,
		"Sort results by name")
	lsCmd.Flags().BoolVar(&flagLsFull, "full", false,
		"Display full values instead of truncating long strings")
}

// LsInteractive is called by the rootCmd when in interactive mode
func lsInteractive(sargs string) {
	args := strings.Split(sargs, " ")
	lsCmd.Flags().Parse(args)
	results := app.GetEntries(app.EntryTypes{Note: true}, flagLsStartsWith, flagLsContains, "", flagLsTags, sortOrder(), flagLsLimit)
	pager := display.NewEntryPager(results)
	pager.PrintPage()
	rl.HistoryDisable()
	rl.SetPrompt("> ")
	for {
		cmd, err := rl.Readline()
		if err != nil {
			fmt.Println("Error:", err)
			break
		} else if num, err := strconv.Atoi(cmd); err == nil {
			ix := num - 1
			if num < 0 || num > len(results.Entries)-1 {
				fmt.Printf("Error: %d is not a valid result number.\n", num)
			} else {
				EntryDetails(pager, results, ix)
				break
			}
		} else if strings.ToLower(cmd) == "n" {
			if !pager.Next() {
				fmt.Println("Error: Already on the last page.")
			}
		} else if strings.ToLower(cmd) == "p" {
			if !pager.Prev() {
				fmt.Println("Error: Already on the first page.")
			}
		} else if cmd == "" || strings.ToLower(cmd) == "q" || strings.ToLower(cmd) == "quit" {
			break
		} else {
			fmt.Println("Error: Unrecognized command:", cmd)
		}
		pager.PrintPage()
	}
	resetLsFlags()
	rl.HistoryEnable()
	rl.SetPrompt(config.Prompt)
}

// EntryDetails displays an entry result in full and provides an entry-specific
// menu of options.
// TODO: Left Off - move this to its own command file and implement
func EntryDetails(pager display.EntryPager, results app.EntryResults, ix int) {
	fmt.Println("Details for", results.Entries[ix].Name())
}

// parseTypes populates an EntryType struct based on the --types flag
func parseTypes() app.EntryTypes {
	types := app.EntryTypes{}
	for _, t := range flagLsTypes {
		switch strings.ToLower(t) {
		case "note", "notes":
			types.Note = true
		case "event", "events":
			types.Event = true
		case "person", "people":
			types.Person = true
		case "place", "places":
			types.Place = true
		case "thing", "things":
			types.Thing = true
		}
	}
	return types
}
