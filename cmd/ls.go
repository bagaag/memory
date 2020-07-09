/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package cmd

import (
	"memory/app"

	"github.com/spf13/cobra"
)

// flag values
var flagStartsWith string
var flagContains string
var flagLimit int = 10
var flagSortModifiedDesc bool
var flagSortName bool
var flagFull bool

// lsCmd lists notes
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Displays and lists entries",
	Long:  `By default, lists 10 most recent entries of any type. Use flags to modify the listing.`,
	Run: func(cmd *cobra.Command, args []string) {
		sort := app.SortRecent
		if flagSortName {
			sort = app.SortName
		}
		//TODO: Implement types flag
		entries := app.GetEntries(app.EntryTypes{Note: true}, flagStartsWith, flagContains, "", []string{}, sort, flagLimit)
		displayEntries(entries, flagFull)
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)

	lsCmd.Flags().StringVar(&flagStartsWith, "starts-with", "",
		"Filter output with case-insensitive prefix")
	lsCmd.Flags().StringVarP(&flagContains, "contains", "c", "",
		"Filter output with case-insensitive substring")
	lsCmd.Flags().IntVarP(&flagLimit, "limit", "l", 10,
		"Specify maximum number of results to return, default is 10")
	lsCmd.Flags().BoolVar(&flagSortModifiedDesc, "sort-modified-desc", false,
		"Sort results with most recently modified entries at the top (default)")
	lsCmd.Flags().BoolVar(&flagSortName, "sort-name", false,
		"Sort results by name")
	lsCmd.Flags().StringVarP(&flagName, "name", "n", "",
		"Display a single note with the given name")
	lsCmd.Flags().BoolVar(&flagFull, "full", false,
		"Display full values instead of truncating long strings")
}
