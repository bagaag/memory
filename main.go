/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/
package main

import (
	"errors"
	"fmt"
	"io"
	"memory/app"
	"memory/app/config"
	"memory/app/persist"
	"memory/cmd/display"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/pkg/term"
	"github.com/urfave/cli"
)

// the rl library provides bash-like completion in interactive mode
var rl *readline.Instance

// inited makes sure we only run cmdInit once
var inited = false

// completer dictates the readline tab completion options
var completer = readline.NewPrefixCompleter(
	readline.PcItem("add-event"),
	readline.PcItem("add-note"),
	readline.PcItem("add-person"),
	readline.PcItem("add-place"),
	readline.PcItem("add-thing"),
	readline.PcItem("detail"),
	readline.PcItem("ls",
		readline.PcItem("--types"),
		readline.PcItem("--tags"),
		readline.PcItem("--contains"),
		readline.PcItem("--start-with"),
	),
	readline.PcItem("delete"),
	readline.PcItem("rename"),
	readline.PcItem("edit"),
)

var cliApp *cli.App

// interactive is true only if program is entered with no sub-command
var interactive = false

func main() {
	cliApp = &cli.App{
		Name:  "memory",
		Usage: `A CLI tool to collect and browse the elements of human experience.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "format",
				Value:    "human",
				Usage:    "how data returned in cli mode is formatted: human or json",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "home",
				Usage:    "directory path where data and settings are read from and saved to",
				Required: false,
			},
		},
		Action: cmdDefault,
		Before: cmdInit,
		Commands: []cli.Command{
			{
				Name:  "add",
				Usage: "adds a new entry",
				Subcommands: []cli.Command{
					{
						Name:   "event",
						Usage:  "adds a new Event entry",
						Action: cmdAdd,
					},
					{
						Name:   "person",
						Usage:  "adds a new Person entry",
						Action: cmdAdd,
					},
					{
						Name:   "place",
						Usage:  "adds a new Place entry",
						Action: cmdAdd,
					},
					{
						Name:   "thing",
						Usage:  "adds a new Thing entry",
						Action: cmdAdd,
					},
					{
						Name:   "note",
						Usage:  "adds a new Note entry",
						Action: cmdAdd,
					},
				},
			},
			{
				Name:   "detail",
				Usage:  "displays details of an entry",
				Action: cmdEdit,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "name of the entry to edit",
						Required: true,
					},
				},
			},
			{
				Name:   "edit",
				Usage:  "edits an entry",
				Action: cmdEdit,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "name of the entry to edit",
						Required: true,
					},
				},
			},
			{
				Name:   "delete",
				Usage:  "deletes an entry",
				Action: cmdDelete,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "name of the entry to delete",
						Required: true,
					},
				},
			},
			{
				Name:   "ls",
				Usage:  "lists entries",
				Action: cmdList,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "contains",
						Usage: "filter on a word or phrase in the name or description",
					},
					&cli.StringFlag{
						Name:  "any-tags",
						Usage: "limit to entries with at least one of these tags, comma-separated",
					},
					&cli.StringFlag{
						Name:  "only-tags",
						Usage: "limit to entries with all of these tags, comma-separated",
					},
					&cli.StringFlag{
						Name:  "order",
						Value: "recent",
						Usage: "order entries by 'recent' or 'name'",
					},
					&cli.IntFlag{
						Name:  "limit",
						Value: -1,
						Usage: "how many entries to return, or -1 for all matching entries",
					},
				},
			},
			{
				Name:   "links",
				Usage:  "displays links to and from an entry",
				Action: cmdLinks,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "name of the entry",
						Required: true,
					},
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(cliApp.Flags))
	sort.Sort(cli.CommandsByName(cliApp.Commands))

	err := cliApp.Run(os.Args)
	if err != nil {
		panic(err)
	}
	rl.Close()
}

// cmdInit runs before any of the cli-invoked cmd functions; exits program on error
var cmdInit = func(c *cli.Context) error {
	if inited {
		return nil
	}
	// init app data
	home := c.String("home")
	if home != "" {
		if !persist.PathExists(home) {
			fmt.Printf("home directory does not exist: %s\n", home)
			os.Exit(1)
		}
		fmt.Printf("Using '%s' as home directory.\n", home)
	}
	err := app.Init(home)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// setup readline
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
	welcomeMessage()
	inited = true
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
	if origEntry.Name != entry.Name {
		app.DeleteEntry(origEntry.Name)
		//TODO: update links on rename
	}
	app.PutEntry(entry)
	app.Save()
	detailInteractiveLoop(entry)
	return nil
}

// cmdDelete deletes an existing entry, identified by name.
var cmdDelete = func(c *cli.Context) error {
	name := c.String("name")
	fmt.Printf("cmdDelete(%s)\n", name)
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
	fmt.Printf("cmdLinks(%s)\n", name)
	return nil
}

// cmdDetail displays details of an entry and, if interactive, provides a menu prompt.
func cmdDetail(c *cli.Context) {
	// --name flag label is optional, "detail 27th Birthday" also works
	var name = c.Args().First()
	if strings.HasPrefix(name, "-") {
		name = c.String("name")
	}
	entry, exists := app.GetEntry(name)
	if !exists {
		fmt.Printf("Entry named '%s' does not exist.\n", name)
	} else {
		detailInteractiveLoop(entry)
	}
}

// filterInput allows certain keys to be intercepted during readline
func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

// welcomeMessage personalizes the app with a message tailored to the visitors current journey.
//TODO: Flesh out the welcome journey
func welcomeMessage() {
	fmt.Printf("Welcome. You have %d entries under management. "+
		"Type 'help' for assistance.\n", app.EntryCount())
}

// parseTypes populates an EntryType struct based on the --types flag
func parseTypes(typesArg []string) app.EntryTypes {
	types := app.EntryTypes{}
	for _, t := range typesArg {
		switch strings.TrimSpace(strings.ToLower(t)) {
		case "note", "notes":
			types.Note = true
		case "event", "events":
			types.Event = true
		case "person", "people":
			types.Person = true
		case "place", "places":
			types.Place = true
		case "thing", "things":
			types.Thing = true
		}
	}
	return types
}

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
			err := cliApp.Run([]string{"memory", "edit", "-name", entry.Name})
			if err != nil {
				fmt.Println("Doh!", err)
			}
		} else if hasLinks && strings.ToLower(cmd) == "l" {
			if !linksInteractiveLoop(entry) {
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
	linkCount := len(entry.LinksTo) + len(entry.LinkedFrom)
	// display links and prompt for command
	display.LinksMenu(entry)
	fmt.Println("\nLinks options: # for details, [b]ack or [Q]uit")
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
		} else if cmd == "" || strings.ToLower(cmd) == "q" {
			return false
		} else {
			fmt.Println("Error: Unrecognized command:", cmd)
		}
	}
}

// getLinkedEntry returns the entry and an 'exists' boolean at the
// given index of an array that combines the LinksTo and LinkedFrom
// slices of the given entry.
func getLinkedEntry(entry app.Entry, ix int) (app.Entry, bool) {
	a := append(entry.LinksTo, entry.LinkedFrom...)
	name := a[ix]
	entry, exists := app.GetEntry(name)
	if !exists {
		fmt.Printf("There is no entry named '%s'.\n", name)
	}
	return entry, exists
}

// Returns either an ascii code, or (if input is an arrow) a Javascript key code.
func readKeyStroke() (ascii int, keyCode int, err error) {
	t, err := term.Open("/dev/tty")
	if err != nil {
		return
	}
	err = term.RawMode(t)
	if err != nil {
		return
	}
	bytes := make([]byte, 3)

	var numRead int
	numRead, err = t.Read(bytes)
	if err != nil {
		return
	}
	if numRead == 3 && bytes[0] == 27 && bytes[1] == 91 {
		// Three-character control sequence, beginning with "ESC-[".

		// Since there are no ASCII codes for arrow keys, we use
		// Javascript key codes.
		if bytes[2] == 65 {
			// Up
			keyCode = 38
		} else if bytes[2] == 66 {
			// Down
			keyCode = 40
		} else if bytes[2] == 67 {
			// Right
			keyCode = 39
		} else if bytes[2] == 68 {
			// Left
			keyCode = 37
		}
	} else if numRead == 1 {
		ascii = int(bytes[0])
	} else {
		// Two characters read??
	}
	t.Restore()
	t.Close()
	fmt.Print("ks:", ascii, keyCode, err, ": ")
	return
}

// editEntry converts an entry to YamlDown, launches an external editor, parses
// the edited content back into an entry and returns the edited entry.
func editEntry(entry app.Entry) (app.Entry, error) {
	initial, err := app.RenderYamlDown(entry)
	if err != nil {
		return app.Entry{}, err
	}
	edited, err := useEditor(initial)
	if err != nil {
		return app.Entry{}, err
	}
	editedEntry, err := app.ParseYamlDown(edited)
	if err != nil {
		return app.Entry{}, err
	}
	return editedEntry, nil
}

// useEditor launches config.editor with a temporary file containing the given string
// waits for the editor to exit and returns a string with the updated content.
func useEditor(s string) (string, error) {
	tmp, err := persist.CreateTempFile(s)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %s", err.Error())
	}
	cmd := exec.Command(config.EditorCommand, tmp)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to interact with editor: %s", err.Error())
	}
	var edited string
	if edited, err = persist.ReadTempFile(tmp); err != nil {
		return "", fmt.Errorf("failed to read temporary file: %s", err.Error())
	}
	if err := persist.RemoveTempFile(tmp); err != nil {
		return edited, fmt.Errorf("failed to delete temporary file: %s", err.Error())
	}
	return edited, nil
}

// Displays prompt for single character input and returns the character entered, or empty string.
func getSingleCharInput() string {
	fmt.Print(config.SubPrompt)
	ascii, _, err := readKeyStroke()
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	s := string(rune(ascii))
	if ascii == 3 { // Ctrl+C
		s = "^C"
	} else if ascii == 13 { // Enter
		s = ""
	}
	fmt.Println(s)
	return s
}
