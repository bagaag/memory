/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/* This file contains code for the main cli flow. */

package cmd

import (
	"sort"

	"github.com/chzyer/readline"
	"github.com/urfave/cli"
)

// the rl library provides bash-like completion in interactive mode
var rl *readline.Instance

// inited makes sure we only run cmdInit once
var inited = false

// completer dictates the readline tab completion options
var completer = readline.NewPrefixCompleter(
	readline.PcItem("add",
		readline.PcItem("event",
			readline.PcItem("-name")),
		readline.PcItem("note",
			readline.PcItem("-name")),
		readline.PcItem("person",
			readline.PcItem("-name")),
		readline.PcItem("place",
			readline.PcItem("-name")),
		readline.PcItem("thing",
			readline.PcItem("-name")),
	),
	readline.PcItem("get",
		readline.PcItem("-name"),
	),
	readline.PcItem("put",
		readline.PcItem("-file"),
	),
	readline.PcItem("detail",
		readline.PcItem("-name"),
	),
	readline.PcItem("ls",
		readline.PcItem("-search"),
		readline.PcItem("-types"),
		readline.PcItem("-tag"),
		readline.PcItem("-any-tag"),
	),
	readline.PcItem("delete",
		readline.PcItem("-name"),
		readline.PcItem("-yes"),
	),
	readline.PcItem("edit",
		readline.PcItem("-name"),
	),
	readline.PcItem("links",
		readline.PcItem("-name"),
	),
	readline.PcItem("seeds"),
	readline.PcItem("rebuild"),
)

var cliApp *cli.App

// interactive is true only if program is entered with no sub-command
var interactive = false

// CreateApp sets up the cli commands and general application flow via the cli lib.
func CreateApp() *cli.App {
	addNameFlag := &cli.StringFlag{
		Name:     "name",
		Usage:    "optional name for the new entry",
		Required: false,
	}
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
				Name:   "add",
				Usage:  "adds a new entry",
				Action: cmdAdd,
				Subcommands: []cli.Command{
					{
						Name:   "event",
						Usage:  "adds a new Event entry",
						Action: cmdAdd,
						Flags:  []cli.Flag{addNameFlag},
					},
					{
						Name:   "person",
						Usage:  "adds a new Person entry",
						Action: cmdAdd,
						Flags:  []cli.Flag{addNameFlag},
					},
					{
						Name:   "place",
						Usage:  "adds a new Place entry",
						Action: cmdAdd,
						Flags:  []cli.Flag{addNameFlag},
					},
					{
						Name:   "thing",
						Usage:  "adds a new Thing entry",
						Action: cmdAdd,
						Flags:  []cli.Flag{addNameFlag},
					},
					{
						Name:   "note",
						Usage:  "adds a new Note entry",
						Action: cmdAdd,
						Flags:  []cli.Flag{addNameFlag},
					},
				},
			},
			{
				Name:   "detail",
				Usage:  "displays details of an entry",
				Action: cmdDetail,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "name of the entry to edit",
						Required: true,
					},
				},
			},
			{
				Name:   "get",
				Usage:  "prints the editable form of an entry",
				Action: cmdGet,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "name of the entry to retrieve",
						Required: true,
					},
				},
			},
			{
				Name:   "put",
				Usage:  "adds or updates an entry from a file",
				Action: cmdPut,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Usage:    "file containing the entry content",
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
					&cli.BoolFlag{
						Name:  "yes",
						Usage: "do not prompt for confirmation",
					},
				},
			},
			{
				Name:   "ls",
				Usage:  "lists entries",
				Action: cmdList,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "search",
						Usage: "search for a word or phrase in the name, tags and description",
					},
					&cli.StringFlag{
						Name:  "tags",
						Usage: "limit to entries with at least one of these tags, comma-separated",
					},
					&cli.StringFlag{
						Name:  "tag",
						Usage: "limit to entries to those with this tag or tags, comma-separated",
					},
					&cli.StringFlag{
						Name:  "types",
						Usage: "comma-separated list of types to list (event, person, place, thing, note)",
					},
					&cli.StringFlag{
						Name:  "order",
						Value: "recent",
						Usage: "order entries by 'recent', 'score' or 'name'",
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
			{
				Name:   "seeds",
				Usage:  "displays links to entries that don't exist yet",
				Action: cmdSeeds,
			},
			{
				Name:   "tags",
				Usage:  "displays summary of entry tags",
				Action: cmdTags,
			},
			{
				Name:   "rebuild",
				Usage:  "rebuilds the search index and internal database from entry files",
				Action: cmdRebuild,
			},
		},
	}

	sort.Sort(cli.FlagsByName(cliApp.Flags))
	sort.Sort(cli.CommandsByName(cliApp.Commands))
	return cliApp
}
