/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/*
This file contains variables and functions used by the
`detail` command in both command line and interactive modes.
*/

package cmd

import (
	"fmt"
	"memory/app"
	"memory/app/config"
	"memory/app/model"
	"memory/cmd/display"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// flag values
var flagDetailType string
var flagDetailName string

// resetFlags returns all flag values to their defaults after being set via
// an interactive command.
func resetDetailFlags() {
	flagDetailName = ""
}

// detailCmd outputs a single entry
var detailCmd = &cobra.Command{
	Use:   "detail",
	Short: "Displays a single entry",
	Long:  `Output details for a single entry based on its unique name.`,
	Run: func(cmd *cobra.Command, args []string) {
		entry, exists := app.GetEntry(flagDetailName)
		if !exists {
			fmt.Printf("Entry named '%s' does not exist.\n", flagDetailName)
		} else {
			display.EntryTable(entry)
		}
	},
}

func init() {
	rootCmd.AddCommand(detailCmd)
	detailCmd.Flags().StringVarP(&flagDetailName, "name", "n", "",
		"The full, case sensitive name of the entry.")
	detailCmd.MarkFlagRequired("name")
}

// detailInteractive is called by the rootCmd when in interactive mode and
// by the lsInteractive command loop.
func detailInteractive(sargs string) {
	// --name flag label is optional, "detail 27th Birthday" also works
	if !strings.HasPrefix(sargs, "-n ") && !strings.HasPrefix(sargs, "--note ") {
		flagDetailName = sargs
	} else {
		lsCmd.Flags().Parse(strings.Split(sargs, " "))
	}
	entry, exists := app.GetEntry(flagDetailName)
	if !exists {
		fmt.Printf("Entry named '%s' does not exist.\n", flagDetailName)
	} else {
		detailInteractiveLoop(entry)
	}
}

// detailInteractiveLoop displays the given entry and prompts for actions
// to take on that entry. Called from the ls interactive loop and from
// detailInteractive.
func detailInteractiveLoop(entry model.IEntry) {
	// setup subloop readline mode
	rl.HistoryDisable()
	rl.SetPrompt(config.SubPrompt)
	defer resetDetailFlags()
	defer rl.HistoryEnable()
	defer rl.SetPrompt(config.Prompt)
	// display detail and prompt for command
	display.EntryTable(entry)
	hasLinks := len(entry.LinksTo())+len(entry.LinkedFrom()) > 0
	if hasLinks {
		fmt.Println("Entry options: [e]dit, [d]elete, [l]inks, [Q]uit")
	} else {
		fmt.Println("Entry options: [e]dit, [d]elete, [Q]uit")
	}
	// interactive loop
	for {
		cmd, err := rl.Readline()
		if err != nil {
			fmt.Println("Error:", err)
			break
		} else if strings.ToLower(cmd) == "e" || strings.ToLower(cmd) == "edit" {
			editInteractive(entry.Name())
			break
		} else if hasLinks && strings.ToLower(cmd) == "l" {
			if !linksInteractiveLoop(entry) {
				break
			}
		} else if cmd == "" || strings.ToLower(cmd) == "q" || strings.ToLower(cmd) == "quit" {
			break
		} else {
			fmt.Println("Error: Unrecognized command:", cmd)
		}
	}
}

// linksInteractiveLoop handles display of an entry's links and
// commands related to them. Returns false if user selects [Q]uit
func linksInteractiveLoop(entry model.IEntry) bool {
	linkCount := len(entry.LinksTo()) + len(entry.LinkedFrom())
	// display links and prompt for command
	display.LinksMenu(entry)
	fmt.Println("\nLinks options: # for details, [b]ack, [Q]uit")
	// interactive loop
	for {
		cmd, err := rl.Readline()
		if err != nil {
			fmt.Println("Error:", err)
			return true
		} else if num, err := strconv.Atoi(cmd); err == nil {
			ix := num - 1
			if ix < 0 || ix >= linkCount {
				fmt.Printf("Error: %d is not a valid link number.\n", num)
			} else {
				nextDetail, exists := getLinkedEntry(entry, ix)
				if exists {
					detailInteractiveLoop(nextDetail)
				}
			}
		} else if strings.ToLower(cmd) == "b" {
			return true
		} else if cmd == "" || strings.ToLower(cmd) == "q" || strings.ToLower(cmd) == "quit" {
			return false
		} else {
			fmt.Println("Error: Unrecognized command:", cmd)
		}
	}
}

// getLinkedEntry returns the entry and an 'exists' boolean at the
// given index of an array that combines the LinksTo and LinkedFrom
// slices of the given entry.
func getLinkedEntry(entry model.IEntry, ix int) (model.IEntry, bool) {
	a := append(entry.LinksTo(), entry.LinkedFrom()...)
	name := a[ix]
	entry, exists := app.GetEntry(name)
	if !exists {
		fmt.Printf("There is no entry named '%s'.\n", name)
	}
	return entry, exists
}
