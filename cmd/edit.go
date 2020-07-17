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
	"memory/app/model"
	"memory/app/util"
	"memory/cmd/display"
	"reflect"

	"github.com/spf13/cobra"
)

// flag values
var flagEditName string
var flagEditNewName string
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
		save := false
		// get the entry we're editing
		entry, exists := app.GetEntry(flagEditName)
		if !exists {
			fmt.Println("Error: Cannot find an entry named", flagEditName)
			return
		}
		// rename if a new-name flag was provided
		if cmd.Flag("new-name").Changed {
			if err := app.RenameEntry(entry.Name(), flagEditNewName); err != nil {
				fmt.Println(util.FormatErrorForDisplay(err))
				return
			}
			// get renamed entry
			entry, exists = app.GetEntry(flagEditNewName)
			if !exists {
				fmt.Println("Error: Cannot find renamed entry ", flagEditName)
				return
			}
			save = true
		}
		switch obj := (entry).(type) {
		case model.Note:
			changed := false
			if cmd.Flag("description").Changed {
				obj.SetDescription(flagEditDescription)
				changed = true
			}
			if cmd.Flag("tags").Changed {
				obj.SetTags(flagEditTags)
				changed = true
			}
			if changed {
				app.PutEntry(obj)
				save = true
			}
		default:
			fmt.Printf("Error: Unhandled entry type: %s\n", reflect.TypeOf(entry))
		}
		if save {
			if err := app.Save(); err != nil {
				fmt.Println("Failed to save data:", err)
				return
			}
		}
		fmt.Printf("Updated note: %s.\n", entry.Name())
		display.EntryTable(entry)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringVarP(&flagEditName, "name", "n", "",
		"Enter a unique name of no more than 50 characters")
	editCmd.Flags().StringVar(&flagEditNewName, "new-name", "",
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
