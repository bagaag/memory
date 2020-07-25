/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/*
This file contains variables and functions used by the
`add-note` command in both command line and interactive modes.
*/

package cmd

import (
	"fmt"
	"memory/app"
	"memory/app/config"
	"memory/app/model"
	"memory/cmd/display"
	"memory/util"

	"github.com/spf13/cobra"
)

// flag values
var flagAddNoteName string
var flagAddNoteDescription string
var flagAddNoteTags []string

// resetFlags returns all flag values to their defaults after being set via
// an interactive command (see addNoteInteractive).
func resetAddNoteFlags() {
	flagAddNoteName = ""
	flagAddNoteDescription = ""
	flagAddNoteTags = []string{}
}

// addNoteCmd adds a new Note
var addNoteCmd = &cobra.Command{
	Use:   "add-note",
	Short: "Adds a new Note",
	Long:  `Adds a new Note. Notes store unstructured information for later use.`,
	Run: func(cmd *cobra.Command, args []string) {
		if flagAddNoteName == "" || len(flagAddNoteName) > config.MaxNameLen {
			fmt.Println("Cannot add note: Name is required and must not exceed 50 characters.")
			return
		}
		note := model.NewNote(flagAddNoteName, flagAddNoteDescription, flagAddNoteTags)
		app.PutEntry(&note)
		if err := app.Save(); err != nil {
			fmt.Println("Failed to save data:", err)
			return
		}
		fmt.Printf("Added note: %s.\n", note.Name())
		display.EntryTable(&note)
	},
}

func init() {
	rootCmd.AddCommand(addNoteCmd)

	addNoteCmd.Flags().StringVarP(&flagAddNoteName, "name", "n", "",
		"Enter a unique name of no more than 50 characters")
	addNoteCmd.Flags().StringVarP(&flagAddNoteDescription, "description", "d", "",
		"Enter a description or omit to launch text editor")
	addNoteCmd.Flags().StringSliceVarP(&flagAddNoteTags, "tags", "t", []string{},
		"Enter comma-separated tags")
}

// addInteractive takes the user through the sequence of prompts to add an item
func addNoteInteractive(sargs string) {
	switch sargs {
	case "note":
		name, err := subPrompt("Enter a name: ", "", validateNoteName)
		if err != nil {
			fmt.Println(util.FormatErrorForDisplay(err))
			return
		}
		desc, err := subPromptEditor("Description", "", "Enter a description: ", emptyValidator)
		if err != nil {
			fmt.Println(util.FormatErrorForDisplay(err))
			return
		}
		tags, err := subPrompt("Enter one or more tags separated by commas: ", "", emptyValidator)
		if err != nil {
			fmt.Println(util.FormatErrorForDisplay(err))
			return
		}
		tagSlice := processTags(tags)
		if name != "" {
			note := model.NewNote(name, desc, tagSlice)
			app.PutEntry(&note)
			app.Save()
			fmt.Println("Note added.")
		}
	default:
		fmt.Printf("%s is not a valid entry type.\n", sargs)
	}
	resetAddNoteFlags()
}
