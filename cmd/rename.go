/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/*
This file contains variables and functions used by the
`rename` command in both command line and interactive modes.
*/

package cmd

import (
	"fmt"
	"memory/app"
	"memory/app/util"

	"github.com/spf13/cobra"
)

// flag values
var flagRenameFrom string
var flagRenameTo string

// resetFlags returns all flag values to their defaults after being set via
// an interactive command
func resetRenameFlags() {
	flagRenameFrom = ""
	flagRenameTo = ""
}

// renameCmd renames an existing entry
var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Renames an entry",
	Long:  `Renames an existing entry.`,
	Run: func(cmd *cobra.Command, args []string) {
		// rename if a new-name flag was provided
		if err := app.RenameEntry(flagRenameFrom, flagRenameTo); err != nil {
			fmt.Println(util.FormatErrorForDisplay(err))
			return
		}
		if err := app.Save(); err != nil {
			fmt.Println("Failed to save data:", err)
			return
		}
		fmt.Printf("Renamed '%s' to '%s'.\n", flagRenameFrom, flagRenameTo)
		resetRenameFlags()
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)

	renameCmd.Flags().StringVar(&flagRenameFrom, "from", "",
		"The name of an existing entry")
	renameCmd.Flags().StringVar(&flagRenameTo, "to", "",
		"A unique name of no more than 50 characters")
	renameCmd.MarkFlagRequired("from")
	renameCmd.MarkFlagRequired("to")
}

// renameEntryInteractive prompts the user for a new entry name
func renameEntryInteractive(name string) {
	entry, exists := app.GetEntry(name)
	if !exists {
		fmt.Println("There is no entry named", name)
		return
	}
	// rename if a new-name flag was provided
	var err error
	flagRenameTo, err = subPrompt("Enter a new name:", name, validateName)
	if err != nil {
		fmt.Println(util.FormatErrorForDisplay(err))
		return
	}
	if err = app.RenameEntry(entry.Name(), flagRenameTo); err != nil {
		fmt.Println(util.FormatErrorForDisplay(err))
		return
	}
	// get renamed entry
	entry, exists = app.GetEntry(name)
	if !exists {
		fmt.Println("Cannot find renamed entry", name)
		return
	}
	// save data
	if err := app.Save(); err != nil {
		fmt.Println("Failed to save data:", err)
		return
	}
	fmt.Printf("Renamed '%s' to '%s'.\n", flagRenameFrom, entry.Name())
	resetRenameFlags()
}
