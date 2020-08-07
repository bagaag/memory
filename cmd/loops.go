/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/* Contains functions that provide interactive input/display loops. */

package cmd

import (
	"fmt"
	"io"
	"memory/app"
	"memory/cmd/display"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
)

// mainLoop provides the main prompt where interactive commands are accepted.
func mainLoop() {
	// input loop
	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF || line == "q" || line == "quit" || line == "exit" {
			break
		}
		line = strings.TrimSpace(line)
		err = cliApp.Run(append([]string{"memory"}, strings.Split(line, " ")...))
		if err != nil {
			fmt.Println("Doh!", err)
		}
	}
	rl.Close()
}

// detailInteractiveLoop displays the given entry and prompts for actions
// to take on that entry. Called from the ls interactive loop and from
// detailInteractive. Returns the bool, true for [b]ack or false for [Q]uit)
func detailInteractiveLoop(entry app.Entry) bool {
	// interactive loop
	for {
		// display detail and prompt for command
		display.EntryTable(entry)
		hasLinks := len(entry.LinksTo)+len(entry.LinkedFrom) > 0
		if hasLinks {
			fmt.Println("Entry options: [e]dit, [d]elete, [l]inks, [b]ack, [Q]uit")
		} else {
			fmt.Println("Entry options: [e]dit, [d]elete, [b]ack, [Q]uit")
		}
		cmd := getSingleCharInput()
		if strings.ToLower(cmd) == "e" {
			// edit entry
			edited, success := editEntryValidationLoop(entry)
			if success {
				entry = edited
			}
		} else if hasLinks && strings.ToLower(cmd) == "l" {
			// display links menu
			if !linksInteractiveLoop(entry) {
				return false
			}
			// update entry in case things changed in the subloops
			var exists bool
			entry, exists = app.GetEntry(entry.Name)
			if !exists {
				return false
			}
		} else if strings.ToLower(cmd) == "d" {
			if deleteEntry(entry.Name, true) {
				return false
			}
		} else if strings.ToLower(cmd) == "b" {
			return true
		} else if cmd == "" || cmd == "^C" || strings.ToLower(cmd) == "q" {
			return false
		} else {
			fmt.Println("Error: Unrecognized command:", cmd)
		}
	}
}

// linksInteractiveLoop handles display of an entry's links and
// commands related to them. Returns false if user selects [B]ack
func linksInteractiveLoop(entry app.Entry) bool {
	// interactive loop
	for {
		linkCount := len(entry.LinksTo) + len(entry.LinkedFrom)
		// display links and prompt for command
		display.LinksMenu(entry)
		fmt.Println("\nLinks options: # for details, [b]ack or [Q]uit")
		cmd := getSingleCharInput()
		if num, err := strconv.Atoi(cmd); err == nil {
			ix := num - 1
			if ix < 0 || ix >= linkCount {
				fmt.Printf("Error: %d is not a valid link number.\n", num)
			} else {
				nextDetail, exists := getLinkedEntry(entry, ix)
				if exists {
					detailInteractiveLoop(nextDetail)
					var exists bool
					entry, exists = app.GetEntry(entry.Name)
					if !exists {
						return false
					}
				}
			}
		} else if strings.ToLower(cmd) == "b" {
			return true
		} else if cmd == "" || cmd == "^C" || strings.ToLower(cmd) == "q" {
			return false
		} else {
			fmt.Println("Error: Unrecognized command:", cmd)
		}
	}
}

// editEntryValidationLoop loads the editor for an entry repeatedly
// until validation passes or the user chooses to discard their edits.
func editEntryValidationLoop(entry app.Entry) (app.Entry, bool) {
	valid := true
	retry := ""
	for {
		var err error
		var edited app.Entry
		edited, retry, err = editEntry(entry, retry)
		_ = retry // eliminates 'retry is declared but not used'
		if err != nil {
			if continueEditingPrompt(err) {
				continue
			} else {
				valid = false
			}
		}
		entry = edited
		break
	}
	return entry, valid
}

// continueEditingPrompt asks the user if they want to continue editing after encountering an error
// preventing a save of the entry after editing.
func continueEditingPrompt(err error) bool {
	fmt.Println("Entry is invalid:", err)
	fmt.Println("Type any key to continue editing or 'd' to discard your changes: ")
	c := getSingleCharInput()
	return c != "d"
}
