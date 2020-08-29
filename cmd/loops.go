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
	"memory/app/model"
	"memory/util"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/mattn/go-shellwords"
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
		// shellwords honors spaces within quotes as a single value, etc.
		args, err := shellwords.Parse(line)
		if err != nil {
			util.FormatErrorForDisplay(err)
			continue
		}
		// prepend "memory" to mimic the args received direclty off the command line
		args = append([]string{"memory"}, args...)
		err = cliApp.Run(args)
		if err != nil {
			fmt.Println(util.FormatErrorForDisplay(err))
		}
	}
	rl.Close()
}

// detailInteractiveLoop displays the given entry and prompts for actions
// to take on that entry. Called from the ls interactive loop and from
// detailInteractive. Returns the bool, true for [b]ack or false for [Q]uit)
func detailInteractiveLoop(entry model.Entry) bool {
	// interactive loop
	for {
		// display detail and prompt for command
		EntryTable(entry)
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
			var err error
			entry, err = memApp.GetEntry(util.GetSlug(entry.Name))
			if err != nil {
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
// commands related to them. Returns true if user selects [B]ack
func linksInteractiveLoop(entry model.Entry) bool {
	// interactive loop
	for {
		links := append(entry.LinksTo, entry.LinkedFrom...)
		linkCount := len(links)
		// display links and prompt for command
		LinksMenu(entry)
		fmt.Println("\nLinks options: # for details, [b]ack or [Q]uit")
		cmd := getSingleCharInput()
		if num, err := strconv.Atoi(cmd); err == nil {
			ix := num - 1
			if ix < 0 || ix >= linkCount {
				fmt.Printf("Error: %d is not a valid link number.\n", num)
			} else {
				linkName := links[ix]
				nextDetail, err := memApp.GetEntry(linkName)
				if err != nil {
					detailInteractiveLoop(nextDetail)
					// refresh entry being inspected after detail loop
					var exists bool
					entry, err = memApp.GetEntry(entry.Slug())
					if !exists {
						return false
					}
				} else {
					if !missingLinkInteractiveLoop(linkName) {
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

// listInteractiveLoop handles the paging of ls results.
func listInteractiveLoop(pager EntryPager) error {
	for {
		input := strings.ToLower(getSingleCharInput())
		if input == "n" {
			if !pager.Next() {
				fmt.Println("Error: Already on the last page.")
			}
		} else if input == "p" {
			if !pager.Prev() {
				fmt.Println("Error: Already on the first page.")
			}
		} else if input == "" || input == "^c" || input == "q" || input == "b" {
			break
		} else if num, err := strconv.Atoi(input); err == nil {
			if num == 0 {
				num = 10
			}
			ix := num - 1
			if ix < 0 || ix >= len(pager.Results.Entries) {
				fmt.Printf("Error: %d is not a valid result number.\n", num)
			} else {
				if !detailInteractiveLoop(pager.Results.Entries[ix]) {
					break
				}
			}
		} else {
			fmt.Println("Error: Unrecognized option:", input)
		}
		pager.PrintPage()
	}
	return nil
}

// editEntryValidationLoop loads the editor for an entry repeatedly
// until validation passes or the user chooses to discard their edits.
func editEntryValidationLoop(entry model.Entry) (model.Entry, bool) {
	valid := true
	retry := ""
	for {
		var err error
		var edited model.Entry
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

// missingLinkInteractiveLoop presents a menu of Entry types to be created
// for the given non-existant entry name. Returns true if [b]ack or false
// if [Q]uit
func missingLinkInteractiveLoop(name string) bool {
	MissingLinkMenu(name)
	types := []string{model.EntryTypeEvent, model.EntryTypePerson, model.EntryTypePlace, model.EntryTypeThing, model.EntryTypeNote}
	for {
		c := getSingleCharInput()
		switch strings.ToLower(c) {
		case "1", "2", "3", "4", "5":
			n, _ := strconv.Atoi(c)
			entryType := types[n-1]
			// add a new entry of the selected type
			args := []string{"memory", "add", strings.ToLower(entryType), "-name", name}
			err := cliApp.Run(args)
			if err != nil {
				fmt.Println(util.FormatErrorForDisplay(err))
			}
			return true
		case "b":
			return true
		case "q":
			return false
		default:
			continue
		}
	}
}
