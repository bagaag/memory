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
	"memory/app"
	"strings"

	"github.com/spf13/cobra"
)

// flag values
var flagLsName string
var flagLsStartsWith string
var flagLsContains string
var flagLsLimit int = 10
var flagLsSortModifiedDesc bool
var flagLsSortName bool
var flagLsFull bool
var flagLsTags []string

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
		//TODO: Implement types flag
		//TODO: Implement name flag
		entries := app.GetEntries(app.EntryTypes{Note: true}, flagLsStartsWith, flagLsContains, "", []string{}, sortOrder(), flagLsLimit)
		displayEntries(entries, flagLsFull)
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
	lsCmd.Flags().StringVar(&flagLsStartsWith, "starts-with", "",
		"Filter output with case-insensitive prefix")
	lsCmd.Flags().StringVarP(&flagLsContains, "contains", "c", "",
		"Filter output with case-insensitive substring")
	lsCmd.Flags().StringSliceVarP(&flagLsTags, "tags", "t", []string{},
		"Limit entries to those tagged with any of these")
	lsCmd.Flags().IntVarP(&flagLsLimit, "limit", "l", 10,
		"Specify maximum number of results to return, default is 10")
	lsCmd.Flags().BoolVar(&flagLsSortModifiedDesc, "sort-modified-desc", false,
		"Sort results with most recently modified entries at the top (default)")
	lsCmd.Flags().BoolVar(&flagLsSortName, "sort-name", false,
		"Sort results by name")
	lsCmd.Flags().StringVarP(&flagLsName, "name", "n", "",
		"Display a single note with the given name")
	lsCmd.Flags().BoolVar(&flagLsFull, "full", false,
		"Display full values instead of truncating long strings")
}

// LsInteractive is called by the rootCmd when in interactive mode
func lsInteractive(sargs string) {
	args := strings.Split(sargs, " ")
	lsCmd.Flags().Parse(args)
	entries := app.GetEntries(app.EntryTypes{Note: true}, flagLsStartsWith, flagLsContains, "", flagLsTags, sortOrder(), flagLsLimit)
	displayEntries(entries, true)
}
