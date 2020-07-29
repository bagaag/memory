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
	"memory/cmd/display"
	"memory/util"
	"strings"

	"github.com/spf13/cobra"
)

// flag values
var flagEditName string
var flagEditField string
var flagEditValue string

// editNoteCmd adds a new Note
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Updates an entry field",
	Long:  `Replaces the value of a named field on an existing entry.`,
	Run: func(cmd *cobra.Command, args []string) {
		flagEditField = strings.ToLower(flagEditField)
		// get the entry we're editing
		entry, exists := app.GetEntry(flagEditName)
		if !exists {
			fmt.Println("Error: Cannot find an entry named", flagEditName)
			return
		}
		// check for 'name' field
		if flagEditField == "name" {
			fmt.Println("Use the rename command to rename an entry.")
			return
		}
		// track if we've found a home for the named field across types
		switch flagEditField {
		case "description":
			entry.Description = flagEditValue
			app.PutEntry(entry)
		case "tags":
			entry.Tags = processTags(flagEditValue)
			app.PutEntry(entry)
		default:
			// error if still no field match
			fmt.Printf("Error: '%s' is not a valid field name for '%s'.", flagEditField, entry.Name)
			return
		}
		// save data
		if err := app.Save(); err != nil {
			fmt.Println("Failed to save data:", err)
			return
		}
		fmt.Printf("Updated entry: %s.\n", entry.Name)
		display.EntryTable(entry)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.Flags().StringVarP(&flagEditName, "name", "n", "",
		"A unique name of no more than 50 characters")
	editCmd.Flags().StringVar(&flagEditField, "field", "",
		"The field name to edit")
	editCmd.Flags().StringVar(&flagEditValue, "value", "",
		"The new value for the field")
	editCmd.MarkFlagRequired("name")
	editCmd.MarkFlagRequired("field")
	editCmd.MarkFlagRequired("value")
}

// editInteractive takes the user through the sequence of prompts to edit an item field or fields
func editInteractive(name string) {
	// get entry and validate name
	entry, exists := app.GetEntry(name)
	if !exists {
		fmt.Println("There is no entry named", name)
		return
	}

	// get list of editable fields based on type
	editableFields := []string{"Edit all fields interactively", "Description", "Tags"}

	// prompt user for which field(s) to edit
	fieldSelection, err := listPrompt("Select a field to edit:", editableFields)
	if err != nil {
		fmt.Println(util.FormatErrorForDisplay(err))
		return
	}

	// update the selected field(s)
	if fieldSelection == 0 || editableFields[fieldSelection] == "Description" {
		desc, err := subPromptEditor("Description", entry.Description, "Enter a description: ", emptyValidator)
		if err != nil {
			fmt.Println(util.FormatErrorForDisplay(err))
			return
		}
		entry.Description = desc

	} else if fieldSelection == 0 || editableFields[fieldSelection] == "Tags" {
		sTags, err := subPrompt("Enter one or more tags separated by commas: ", strings.Join(entry.Tags, ","), emptyValidator)
		if err != nil {
			fmt.Println(util.FormatErrorForDisplay(err))
			return
		}
		tags := processTags(sTags)
		entry.Tags = tags
	}
	// update entry in collection
	app.PutEntry(entry)
	// save data
	if err := app.Save(); err != nil {
		fmt.Println("Failed to save data:", err)
		return
	}
	fmt.Println("Entry updated.")
	display.EntryTable(entry)
}
