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
	"os"
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
				os.Exit(0)
			} else {
				continue
			}
		} else if err == io.EOF || line == "q" || line == "quit" || line == "exit" {
			break
		}
		mainLoopInput = line
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
		entryLinks, _ := memApp.Search.Links(entry.Slug())
		reverseLinks, _ := memApp.Search.ReverseLinks(entry.Slug())
		hasLinks := len(entryLinks)+len(reverseLinks) > 0
		optionalCommands := ""
		if hasLinks {
			optionalCommands = ", [l]inks"
		}
		fmt.Println("Entry options: [e]dit, [d]elete" + optionalCommands + ", [a]ttachments, [b]ack, [Q]uit")
		cmd := getSingleCharInput()
		updateEntry := false // set to true if the update may have changed due to a sub-command
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
			updateEntry = true
		} else if strings.ToLower(cmd) == "a" {
			if !filesInteractiveLoop(entry) {
				return false
			}
			updateEntry = true
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
		// update entry in case things changed in the subloops
		if updateEntry {
			var err error
			entry, err = memApp.GetEntry(util.GetSlug(entry.Name))
			if err != nil {
				return false
			}
			updateEntry = false
		}
	}
}

// filesInteractiveLoop handles display of an entry's files and
// commands related to them. Returns true if user selects [B]ack
func filesInteractiveLoop(entry model.Entry) bool {
	if !entry.Populated() {
		var err error
		if entry, err = memApp.GetEntry(entry.Slug()); err != nil {
			fmt.Println(util.FormatErrorForDisplay(err))
			return false
		}
	}
	// interactive loop
	for {
		// display links and prompt for command
		FilesMenu(entry)
		detailCmd := ""
		if len(entry.Attachments) > 0 {
			detailCmd = "# for details, "
		}
		fmt.Println("\nAttachment options: " + detailCmd + "[a]dd, [b]ack or [Q]uit")
		cmd := getSingleCharInput()
		lcmd := strings.ToLower(cmd)
		if num, err := strconv.Atoi(cmd); err == nil {
			ix := num - 1
			if ix < 0 || ix >= len(entry.Attachments) {
				fmt.Printf("Error: %d is not a valid attachment number.\n", num)
			} else {
				fileInteractiveLoop(entry, ix)
			}
		} else if lcmd == "a" {
			args := []string{"memory", "file", "add", "-entry", entry.Slug()}
			err = cliApp.Run(args)
			if err != nil {
				fmt.Println(util.FormatErrorForDisplay(err))
			} else {
				fmt.Println("Attachment added.")
				entry, _ = memApp.GetEntry(entry.Slug())
			}
		} else if lcmd == "b" {
			return true
		} else if cmd == "" || cmd == "^C" || lcmd == "q" {
			return false
		} else {
			fmt.Println("Error: Unrecognized command:", cmd)
		}
		// refresh entry before looping
		entry, _ = memApp.GetEntry(entry.Slug())
	}
}

// fileInteractiveLoop handles display of an attachment and
// commands related to it. Returns true if user selects [B]ack
func fileInteractiveLoop(entry model.Entry, ix int) bool {
	att := entry.Attachments[ix]
	// interactive loop
	for {
		// display links and prompt for command
		fmt.Println("\nAttachment: " + att.Name + " [" + att.DisplayFileName() + "]\n")
		fmt.Println("Options: [o]pen, [r]ename, [d]elete, [b]ack or [Q]uit")
		cmd := getSingleCharInput()
		lcmd := strings.ToLower(cmd)
		if lcmd == "o" {
			// open command
			args := []string{"memory", "file", "open",
				"-entry", entry.Slug(),
				"-title", att.Name}
			if err := cliApp.Run(args); err != nil {
				fmt.Println(util.FormatErrorForDisplay(err))
			}
		} else if lcmd == "r" {
			// rename command
			newTitle, err := subPrompt("Enter a new name for the attachment: ", att.Name, emptyValidator)
			if err != nil {
				fmt.Println(util.FormatErrorForDisplay(err))
				continue
			}
			args := []string{"memory", "file", "rename", "" +
				"-entry", entry.Slug(),
				"-title", att.Name,
				"-new-title", newTitle}
			if err := cliApp.Run(args); err != nil {
				fmt.Println(util.FormatErrorForDisplay(err))
			}
			return true
		} else if lcmd == "d" {
			// delete command
			answer, err := subPrompt("Are you sure you want to delete this attachment? [y,N]: ", "", validateYesNo)
			if err != nil {
				fmt.Print(util.FormatErrorForDisplay(err))
				continue
			}
			if answer != "y" {
				continue
			}
			args := []string{"memory", "file", "delete", "" +
				"-entry", entry.Slug(),
				"-title", att.Name}
			if err := cliApp.Run(args); err != nil {
				fmt.Println(util.FormatErrorForDisplay(err))
			}
			return true
		} else if lcmd == "b" {
			return true
		} else if cmd == "" || cmd == "^C" || lcmd == "q" {
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
		slug := entry.Slug()
		entryLinks, _ := memApp.Search.Links(slug)
		reverseLinks, _ := memApp.Search.ReverseLinks(slug)
		linkCount := len(entryLinks) + len(reverseLinks)
		// display links and prompt for command
		LinksMenu(entry)
		fmt.Println("\nLinks options: # for details, [b]ack or [Q]uit")
		cmd := getSingleCharInput()
		if num, err := strconv.Atoi(cmd); err == nil {
			ix := num - 1
			if ix < 0 || ix >= linkCount {
				fmt.Printf("Error: %d is not a valid link number.\n", num)
			} else {
				linkName := entryLinks[ix]
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
				entry, err := memApp.GetEntry(pager.Results.Entries[ix].Slug())
				if err != nil {
					return err
				}
				if !detailInteractiveLoop(entry) {
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
