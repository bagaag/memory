/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
  "memory/app"
)

// adds a new Note 
var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Adds a new Note",
	Long: `Adds a new Note. Notes store unstructured information for later use.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("note -d %s -t %s\n", cmd.Flag("description").Value, cmd.Flag("tags").Value)
    note := app.NewNote("hey now", []string{"one","two"})
    app.PutNote(note)
	},
}

func init() {
	rootCmd.AddCommand(noteCmd)
  noteCmd.Flags().StringP("description", "d", "", "Enter a description or omit to launch text editor")
  noteCmd.Flags().StringSliceP("tags", "t", []string{}, "Enter comma-separated tags")
  //noteCmd.MarkFlagRequired("description")
}
