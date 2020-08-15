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
	if len(c.Args()) == 0 {
		// say hi if we're in interactive mode
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
	var entry app.Entry
	var success = false
	// validate entry type
	entryType := strings.Title(c.Command.Name)
	if entryType == "" {
		return errors.New("missing entry type: [event, person, place, thing, note]")
	}
	// display editor w/ template if no file is provided
	newEntry := app.NewEntry(entryType, "New "+entryType, "", []string{})
	entry, success = editEntryValidationLoop(newEntry)
	if !success {
		return errors.New("failed to add a valid entry")
	}
	app.PutEntry(entry)
	app.Save()
	app.UpdateLinks()
	fmt.Println("Added new entry:", entry.Name)
	display.EntryTable(entry)
	return nil
}

// cmdPut adds or updates an entry from the given file.
var cmdPut = func(c *cli.Context) error {
	// read from file if -file is provided
	content, err := persist.ReadFile(c.String("file"))
	if err != nil {
		return err
	}
	entry, err := parseEntryText(content)
	if err != nil {
		return err
	}
	_, existed := app.GetEntry(entry.Slug())
	app.PutEntry(entry)
	app.Save()
	if existed {
		fmt.Println("Updated entry:", entry.Name)
	} else {
		fmt.Println("Added new entry:", entry.Name)
	}
	display.EntryTable(entry)
	return nil
}

// cmdEdit edits an existing entry, identified by name.
var cmdEdit = func(c *cli.Context) error {
	name := c.String("name")
	origEntry, exists := app.GetEntryByName(name)
	if !exists {
		return fmt.Errorf("there is no entry named '%s'", name)
	}
	entry, success := editEntryValidationLoop(origEntry)
	if !success {
		return errors.New("failed to add a valid entry")
	}
	if origEntry.Name != entry.Name {
		// entry being renamed
		_, exists := app.GetEntryByName(entry.Name)
		if exists {
			return errors.New("cannot rename entry; an entry with this name already exists")
		}
		app.DeleteEntry(app.GetSlug(origEntry.Name))
	}
	app.PutEntry(entry)
	app.Save()
	app.UpdateLinks()
	fmt.Println("Updated entry:", entry.Name)
	display.EntryTable(entry)
	return nil
}

// cmdDelete deletes an existing entry, identified by name.
var cmdDelete = func(c *cli.Context) error {
	name := c.String("name")
	ask := !c.Bool("yes")
	deleteEntry(name, ask)
	app.UpdateLinks()
	return nil
}

// cmdList lists entries, optionally filtered and sorted.
var cmdList = func(c *cli.Context) error {
	contains := c.String("contains")
	anyTags := []string{}
	if c.IsSet("tags") {
		anyTags = strings.Split(c.String("any-tags"), ",")
	}
	onlyTags := []string{}
	if c.IsSet("tag") {
		onlyTags = strings.Split(c.String("tag"), ",")
	}
	order := app.SortRecent
	if c.String("order") == "name" {
		order = app.SortName
	}
	limit := c.Int("limit")
	types := strings.Split(c.String("types"), "")
	startsWith := "" //TODO: Implement or remove ls startsWith
	search := ""     //TODO: Implement or remove ls search

	results := app.GetEntries(parseTypes(types), startsWith, contains, search, onlyTags, anyTags, order, limit)

	if interactive {
		pager := display.NewEntryPager(results)
		pager.PrintPage()
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
	entry, exists := app.GetEntryByName(name)
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

// cmdGet displays the editable content of an entry
func cmdGet(c *cli.Context) error {
	name := c.String("name")
	entry, exists := app.GetEntry(app.GetSlug(name))
	if !exists {
		return nil
	}
	content, err := app.RenderYamlDown(entry)
	if err != nil {
		return err
	}
	fmt.Println(content)
	return nil
}

// cmdDetail displays details of an entry and, if interactive, provides a menu prompt.
func cmdDetail(c *cli.Context) {
	name := c.String("name")
	entry, exists := app.GetEntry(app.GetSlug(name))
	if !exists {
		fmt.Printf("Entry named '%s' does not exist.\n", name)
	} else if interactive {
		detailInteractiveLoop(entry)
	} else {
		display.EntryTable(entry)
	}
}

// cmdTags displays a list of tags in use and how many entries each has
func cmdTags(c *cli.Context) {
	tags := app.GetTags()
	sorted := app.GetSortedTags(tags)
	fmt.Println()
	for _, tag := range sorted {
		names := tags[tag]
		fmt.Printf("%s [%d]  ", tag, len(names))
	}
	fmt.Println()
	fmt.Println()
}
