/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
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
var flagName string
var flagDescription string
var flagTags []string

// addNoteCmd adds a new Note
var addNoteCmd = &cobra.Command{
	Use:   "add-note",
	Short: "Adds a new Note",
	Long:  `Adds a new Note. Notes store unstructured information for later use.`,
	Run: func(cmd *cobra.Command, args []string) {
		if flagName == "" || len(flagName) > config.MaxNameLen {
			fmt.Println("Cannot add note: Name is required and must not exceed 50 characters.")
			return
		}
		note := model.NewNote(flagName, flagDescription, flagTags)
		app.PutNote(note)
		save()
		fmt.Printf("Added note: %s.\n", note.Name())
		//TODO: Display new note
	},
}

func init() {
	rootCmd.AddCommand(addNoteCmd)

	addNoteCmd.Flags().StringVarP(&flagName, "name", "n", "",
		"Enter a unique name of no more than 50 characters")
	addNoteCmd.Flags().StringVarP(&flagDescription, "description", "d", "",
		"Enter a description or omit to launch text editor")
	addNoteCmd.Flags().StringSliceVarP(&flagTags, "tags", "t", []string{},
		"Enter comma-separated tags")
}
