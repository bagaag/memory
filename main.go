/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"memory/app"
	"memory/app/config"
	"memory/cmd/display"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/pkg/term"
	"github.com/urfave/cli"
)

// the rl library provides bash-like completion in interactive mode
var rl *readline.Instance

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
var interactive = -1

func main() {

	cliApp = &cli.App{
		Name:  "memory",
		Usage: `A CLI tool to collect and browse the elements of human experience.`,
		//TODO: Figure out settings file
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "format",
				Value:    "human",
				Usage:    "how data returned in cli mode is formatted: human or json",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "editor",
				Value:    "/usr/bin/vim",
				Usage:    "command to invoke when editing an entry",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "data-file",
				Value:    "~/.memory.db",
				Usage:    "path to data file where entries are read from and saved to",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "interactive",
				Usage:    "enter an interactive session after completing command",
				Required: false,
			},
		},
		Action: cmdDefault,
		Commands: []cli.Command{
			{
				Name:   "add",
				Usage:  "adds a new entry",
				Action: cmdAdd,
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
					&cli.BoolFlag{
						Name:  "no-paging",
						Usage: "disable interactive paging of results",
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
		log.Fatal(err)
	}
}

// setInteractive establishes whether the app was entered in interactive
// mode or if --interactive was included with a command. If interactive
// is false, the program exits after completing an initial command.
func setInteractive(subCommand bool, interactiveFlag bool) {
	if interactive > -1 {
		return
	}
	if subCommand && interactiveFlag {
		interactive = 1
	} else {
		interactive = 0
	}
}

// cmdDefault command enters the interactive command loop.
var cmdDefault = func(c *cli.Context) error {
	setInteractive(false, c.Bool("interactive"))
	mainLoop()
	return nil
}

// cmdAdd adds a new entry. Requires a sub-command indicating type.
var cmdAdd = func(c *cli.Context) error {
	setInteractive(true, c.Bool("interactive"))
	t := c.Command.Name
	if t == "" {
		fmt.Println("USAGE:\n   add [event, person, place, thing, note]")
	}
	fmt.Println(t)
	return nil
}

// cmdEdit edits an existing entry, identified by name.
var cmdEdit = func(c *cli.Context) error {
	setInteractive(true, c.Bool("interactive"))
	name := c.String("name")
	fmt.Printf("cmdEdit(%s)\n", name)
	return nil
}

// cmdDelete deletes an existing entry, identified by name.
var cmdDelete = func(c *cli.Context) error {
	setInteractive(true, c.Bool("interactive"))
	name := c.String("name")
	fmt.Printf("cmdDelete(%s)\n", name)
	return nil
}

// cmdList lists entries, optionally filtered and sorted.
var cmdList = func(c *cli.Context) error {
	setInteractive(true, c.Bool("interactive"))
	contains := c.String("contains")
	anyTags := strings.Split(c.String("any-tags"), ",")
	//onlyTags := strings.Split(c.String("only-tags"), ",") //TODO: Implement onlyTags ls option
	noPaging := c.Bool("no-paging")
	order := app.SortRecent
	if c.String("order") == "name" {
		order = app.SortName
	}
	limit := c.Int("limit")
	types := strings.Split(c.String("types"), "")
	startsWith := "" //TODO: Implement or remove ls startsWith
	search := ""     //TODO: Implement or remove ls search

	results := app.GetEntries(parseTypes(types), startsWith, contains, search, anyTags, order, limit)
	if !noPaging {

		pager := display.NewEntryPager(results)
		pager.PrintPage()
		//rl.HistoryDisable()
		//rl.SetPrompt(config.SubPrompt)
		//defer rl.HistoryEnable()
		//defer rl.SetPrompt(config.Prompt)
		for {
			//char, err := readChar()
			ascii, _, err := getChar()
			s := string(rune(ascii))
			if err != nil {
				fmt.Println("Error:", err)
				break
			} else if num, err := strconv.Atoi(s); err == nil {
				ix := num - 1
				if ix < 0 || ix >= len(results.Entries) {
					fmt.Printf("Error: %d is not a valid result number.\n", num)
				} else {
					detailInteractiveLoop(results.Entries[ix])
					break
				}
			} else if strings.ToLower(s) == "n" {
				if !pager.Next() {
					fmt.Println("Error: Already on the last page.")
				}
			} else if strings.ToLower(s) == "p" {
				if !pager.Prev() {
					fmt.Println("Error: Already on the first page.")
				}
			} else if s == "" || strings.ToLower(s) == "q" {
				break
			} else {
				fmt.Println("Error: Unrecognized option:", s)
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
	setInteractive(true, c.Bool("interactive"))
	name := c.String("name")
	fmt.Printf("cmdLinks(%s)\n", name)
	return nil
}

// cmdDetail displays details of an entry and, if interactive, provides a menu prompt.
func cmdDetail(c *cli.Context) {
	setInteractive(true, c.Bool("interactive"))
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

// readChar reads a single keystroke from the console
func readChar() (string, error) {
	//reader := bufio.NewReader(os.Stdin)
	//char, _, err := reader.ReadRune()
	//return char, err
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text(), nil
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
	welcomeMessage()
	// readline setup
	var err error
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
	defer rl.Close()
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
		cliApp.Run(append([]string{"memory"}, strings.Split(line, " ")...))
	}
}

// detailInteractiveLoop displays the given entry and prompts for actions
// to take on that entry. Called from the ls interactive loop and from
// detailInteractive.
func detailInteractiveLoop(entry app.Entry) {
	// setup subloop readline mode
	rl.HistoryDisable()
	rl.SetPrompt(config.SubPrompt)
	defer rl.HistoryEnable()
	defer rl.SetPrompt(config.Prompt)
	// display detail and prompt for command
	display.EntryTable(entry)
	hasLinks := len(entry.LinksTo)+len(entry.LinkedFrom) > 0
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
			cliApp.Run([]string{"memory", "edit", entry.Name})
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
func linksInteractiveLoop(entry app.Entry) bool {
	linkCount := len(entry.LinksTo) + len(entry.LinkedFrom)
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
func getChar() (ascii int, keyCode int, err error) {
	t, _ := term.Open("/dev/tty")
	term.RawMode(t)
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
	return
}
