/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/*
This file contains variables and functions used by the
`delete` command in both command line and interactive modes.
*/

package cmd

import (
	"fmt"
	"memory/app"
	"memory/cmd/display"
	"memory/util"
	"strings"

	"github.com/spf13/cobra"
)

// flag values
var flagDeleteName string

// resetFlags returns all flag values to their defaults after being set via
// an interactive command
func resetDeleteFlags() {
	flagDeleteName = ""
}

// deleteCmd removes an existing entry from the collection
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes an entry",
	Long:  `Deletes an existing entry.`,
	Run: func(cmd *cobra.Command, args []string) {
		deleteEntry(flagDeleteName)
		resetRenameFlags()
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVarP(&flagDeleteName, "name", "n", "",
		"The name of an existing entry to delete")
	deleteCmd.MarkFlagRequired("name")
}

func deleteEntry(name string) {
	success := app.DeleteEntry(name)
	if !success {
		fmt.Printf("Entry named '%s' does not exist.\n", name)
	} else {
		fmt.Printf("Deleted entry named '%s'.\n", name)
	}
	// save data
	if err := app.Save(); err != nil {
		fmt.Println("Failed to save data:", err)
		return
	}
}

// deleteEntryInteractive displays entry details, and prompts for confirmation before deleting
func deleteEntryInteractive(name string) {
	entry, exists := app.GetEntry(name)
	if !exists {
		fmt.Printf("There is no entry named %s.\n", name)
		return
	}
	display.EntryTable(entry)
	confirm, err := subPrompt("Are you sure you want to delete this entry? [N/y]: ", "", validateYesNo)
	if err != nil {
		fmt.Println(util.FormatErrorForDisplay(err))
		return
	}
	if strings.ToLower(confirm) == "y" {
		deleteEntry(name)
	}
}
