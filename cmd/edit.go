/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/*
This file contains variables and functions used by the
`edit-note` command in both command line and interactive modes.
*/

package cmd

import (
	"fmt"
	"memory/app"
	"memory/app/config"
	"memory/app/model"

	"github.com/spf13/cobra"
)

// flag values
var flagEditName string
var flagEditDescription string
var flagEditTags []string

// resetFlags returns all flag values to their defaults after being set via
// an interactive command (see editNoteInteractive).
func resetEditFlags() {
	flagEditName = ""
	flagEditDescription = ""
	flagEditTags = []string{}
}

// editNoteCmd adds a new Note
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edits an entry",
	Long:  `Edits an existing entry.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(flagEditName) > config.MaxNameLen {
			fmt.Println("Name cannot exceed 50 characters.")
			return
		}

		note := model.NewNote(flagEditName, flagEditDescription, flagEditTags)
		app.PutEntry(note)
		save()
		fmt.Printf("Added note: %s.\n", note.Name())
		//TODO: Display new note
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringVarP(&flagEditName, "name", "n", "",
		"Enter a unique name of no more than 50 characters")
	editCmd.Flags().StringVar(&flagEditName, "new-name", "",
		"Enter a unique name of no more than 50 characters")
	editCmd.Flags().StringVarP(&flagEditDescription, "description", "d", "",
		"Enter a description or omit to launch text editor")
	editCmd.Flags().StringSliceVarP(&flagEditTags, "tags", "t", []string{},
		"Enter comma-separated tags")
	editCmd.MarkFlagRequired("name")
}

// addInteractive takes the user through the sequence of prompts to add an item
func editNoteInteractive(sargs string) {
	switch sargs {
	case "note":
		name := subPrompt("Enter a name: ", validateNoteName)
		desc := subPromptEditor("Description", "", "Enter a description: ", emptyValidator)
		tags := subPrompt("Enter one or more tags separated by commas: ", emptyValidator)
		tagSlice := processTags(tags)
		if name != "" {
			note := model.NewNote(name, desc, tagSlice)
			app.PutEntry(note)
			app.Save()
			fmt.Println("Note added.")
		}
	default:
		fmt.Printf("%s is not a valid entry type.\n", sargs)
	}
	resetEditFlags()
}
