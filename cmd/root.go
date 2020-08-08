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
		readline.PcItem("event"),
		readline.PcItem("note"),
		readline.PcItem("person"),
		readline.PcItem("place"),
		readline.PcItem("thing"),
	),
	readline.PcItem("detail",
		readline.PcItem("-name"),
	),
	readline.PcItem("ls",
		readline.PcItem("-types"),
		readline.PcItem("-tags"),
		readline.PcItem("-contains"),
		readline.PcItem("-start-with"),
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
)

var cliApp *cli.App

// interactive is true only if program is entered with no sub-command
var interactive = false

// CreateApp sets up the cli commands and general application flow via the cli lib.
func CreateApp() *cli.App {
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
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "file",
						Usage: "file containing the entry content",
					},
				},
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
				Name:   "edit",
				Usage:  "edits an entry",
				Action: cmdEdit,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "name of the entry to edit",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "file",
						Usage: "file containing the entry content",
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
						Name:  "contains",
						Usage: "filter on a word or phrase in the name or description",
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
			{
				Name:   "tags",
				Usage:  "displays summary of entry tags",
				Action: cmdTags,
			},
		},
	}

	sort.Sort(cli.FlagsByName(cliApp.Flags))
	sort.Sort(cli.CommandsByName(cliApp.Commands))
	return cliApp
}
