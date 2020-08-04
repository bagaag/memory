/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/* This file contains the cmd* command functions invoked by the cli loop in root.go. */

package cmd

import (
	"errors"
	"fmt"
	"memory/app"
	"memory/app/config"
	"memory/app/persist"
	"memory/cmd/display"
	"os"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/urfave/cli"
)

// cmdInit runs before any of the cli-invoked cmd functions; exits program on error
var cmdInit = func(c *cli.Context) error {
	if inited {
		return nil
	}
	// init app data
	home := c.String("home")
	if home != "" {
		if !persist.PathExists(home) {
			fmt.Printf("Error: Home directory does not exist: %s\n", home)
			os.Exit(1)
		}
		fmt.Printf("Using '%s' as home directory.\n", home)
	}
	err := app.Init(home)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if len(c.Args()) == 0 {
		// setup readline if we're going to be interactive
		rl, err = readline.NewEx(&readline.Config{
			Prompt:              config.Prompt,
			HistoryFile:         config.HistoryPath(),
			AutoComplete:        completer,
			InterruptPrompt:     "^C",
			EOFPrompt:           "exit",
			HistorySearchFold:   true,
			FuncFilterInputRune: filterInput,
		})
		if err != nil {
			panic(err)
		}
		// say hi
		display.WelcomeMessage()
		inited = true
	}
	return nil
}

// cmdDefault command enters the interactive command loop.
var cmdDefault = func(c *cli.Context) error {
	interactive = true
	mainLoop()
	return nil
}

// cmdAdd adds a new entry. Requires a sub-command indicating type.
var cmdAdd = func(c *cli.Context) error {
	entryType := strings.Title(c.Command.Name)
	if entryType == "" {
		return errors.New("missing entry type: [event, person, place, thing, note]")
	}
	name := c.String("name")
	newEntry := app.NewEntry(entryType, name, "", []string{})
	entry, err := editEntry(newEntry)
	if err != nil {
		return err
	}
	app.PutEntry(entry)
	app.Save()
	fmt.Println("Added new entry:", entry.Name)
	display.EntryTable(entry)
	return nil
}

// cmdEdit edits an existing entry, identified by name.
var cmdEdit = func(c *cli.Context) error {
	name := c.String("name")
	origEntry, exists := app.GetEntry(name)
	if !exists {
		return fmt.Errorf("there is no entry named '%s'", name)
	}
	entry, err := editEntry(origEntry)
	if err != nil {
		return err
	}
	detailInteractiveLoop(entry)
	return nil
}

// cmdDelete deletes an existing entry, identified by name.
var cmdDelete = func(c *cli.Context) error {
	name := c.String("name")
	ask := !c.Bool("yes")
	deleteEntry(name, ask)
	return nil
}

// cmdList lists entries, optionally filtered and sorted.
var cmdList = func(c *cli.Context) error {
	contains := c.String("contains")
	anyTags := []string{}
	if c.IsSet("any-tags") {
		anyTags = strings.Split(c.String("any-tags"), ",")
	}
	//onlyTags := strings.Split(c.String("only-tags"), ",") //TODO: Implement onlyTags ls option
	order := app.SortRecent
	if c.String("order") == "name" {
		order = app.SortName
	}
	limit := c.Int("limit")
	types := strings.Split(c.String("types"), "")
	startsWith := "" //TODO: Implement or remove ls startsWith
	search := ""     //TODO: Implement or remove ls search

	results := app.GetEntries(parseTypes(types), startsWith, contains, search, anyTags, order, limit)

	if interactive {
		pager := display.NewEntryPager(results)
		pager.PrintPage()
		for {
			input := getSingleCharInput()
			if strings.ToLower(input) == "n" {
				if !pager.Next() {
					fmt.Println("Error: Already on the last page.")
				}
			} else if strings.ToLower(input) == "p" {
				if !pager.Prev() {
					fmt.Println("Error: Already on the first page.")
				}
			} else if input == "" || input == "^C" || strings.ToLower(input) == "q" || strings.ToLower(input) == "b" {
				break
			} else if num, err := strconv.Atoi(input); err == nil {
				ix := num - 1
				if ix < 0 || ix >= len(results.Entries) {
					fmt.Printf("Error: %d is not a valid result number.\n", num)
				} else {
					if !detailInteractiveLoop(results.Entries[ix]) {
						break
					}
					results = app.RefreshResults(results)
					pager = display.NewEntryPager(results)
				}
			} else {
				fmt.Println("Error: Unrecognized option:", input)
			}
			pager.PrintPage()
		}
	} else {
		display.EntryTables(results.Entries)
	}
	return nil
}

// cmdLinks lists the entries linked to and from an existing entry, identified by name.
var cmdLinks = func(c *cli.Context) error {
	name := c.String("name")
	entry, exists := app.GetEntry(name)
	if !exists {
		fmt.Println("Cannot find entry named", name)
		return errors.New("entry not found")
	}
	if interactive {
		linksInteractiveLoop(entry)
	} else {
		display.LinksMenu(entry)
		fmt.Println("")
	}
	return nil
}

// cmdDetail displays details of an entry and, if interactive, provides a menu prompt.
func cmdDetail(c *cli.Context) {
	name := c.String("name")
	entry, exists := app.GetEntry(name)
	if !exists {
		fmt.Printf("Entry named '%s' does not exist.\n", name)
	} else if interactive {
		detailInteractiveLoop(entry)
	} else {
		display.EntryTable(entry)
	}
}
