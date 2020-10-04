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
	"github.com/chzyer/readline"
	"github.com/urfave/cli"
	"memory/app/config"
	"memory/app/links"
	"memory/app/localfs"
	"memory/app/memory"
	"memory/app/model"
	"memory/app/search"
	"memory/app/template"
	"memory/util"
	"os"
	"os/exec"
	"strings"
	"time"
)

// cmdInit runs before any of the cli-invoked cmd functions; exits program on error
func cmdInit(c *cli.Context) error {
	if inited {
		return nil
	}
	// init app data
	home := c.String("home")
	if home != "" {
		if !localfs.PathExists(home) {
			fmt.Printf("Error: Home directory does not exist: %s\n", home)
			os.Exit(1)
		}
		fmt.Printf("Using '%s' as home directory.\n", home)
	}
	var err error
	// initialize Memory app object
	memApp, err = memory.Init(home)
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
		WelcomeMessage()
		inited = true
	}
	return nil
}

// cmdDefault command enters the interactive command loop.
func cmdDefault(c *cli.Context) error {
	if len(c.Args()) > 0 && firstCommand {
		cli.ShowAppHelpAndExit(c, 1)
	} else {
		firstCommand = false
		if mainLoopInput != "" {
			fmt.Println("Not sure what to do with that. Try 'help'.")
		}
	}
	interactive = true
	mainLoop()
	return nil
}

// cmdAdd adds a new entry. Requires a sub-command indicating type.
func cmdAdd(c *cli.Context) error {
	var entry model.Entry
	var success = false
	// validate entry type
	entryType := strings.Title(c.Command.Name)
	if entryType == "" {
		return errors.New("missing entry type: [event, person, place, thing, note]")
	}
	// display editor w/ template if no file is provided
	name := "New " + entryType
	if c.IsSet("name") {
		name = c.String("name")
	}
	newEntry := model.NewEntry(entryType, name, "", []string{})
	entry, success = editEntryValidationLoop(newEntry)
	if !success {
		return errors.New("failed to add a valid entry")
	}
	fmt.Println("Added new entry:", entry.Name)
	EntryTable(entry)
	return nil
}

// cmdPut adds or updates an entry from the given file.
func cmdPut(c *cli.Context) error {
	// read from file if -file is provided
	// TODO: support .md/txt and .json
	content, _, err := localfs.ReadFile(c.String("file"))
	if err != nil {
		return err
	}
	entry, err := parseEntryText(content)
	if err != nil {
		return err
	}
	existed := memApp.EntryExists(entry.Slug())
	entry.Modified = time.Now()
	if !existed {
		entry.Created = entry.Modified
	}
	if err := memApp.PutEntry(entry); err != nil {
		return err
	}
	if existed {
		fmt.Println("Updated entry:", entry.Name)
	} else {
		fmt.Println("Added new entry:", entry.Name)
	}
	EntryTable(entry)
	return nil
}

// cmdEdit edits an existing entry, identified by name.
func cmdEdit(c *cli.Context) error {
	name := c.String("name")
	origEntry, err := memApp.GetEntry(util.GetSlug(name))
	origEntry.Description = links.RenderLinks(origEntry.Description, memApp.EntryExists)
	if model.IsEntryNotFound(err) {
		return fmt.Errorf("there is no entry named '%s'", name)
	} else if err != nil {
		return err
	}
	entry, success := editEntryValidationLoop(origEntry)
	if !success {
		return errors.New("failed to add a valid entry")
	}
	if origEntry.Name != entry.Name {
		if entry, err = memApp.RenameEntry(origEntry.Name, entry.Name); err != nil {
			return err
		}
	}
	if err := memApp.PutEntry(entry); err != nil {
		return err
	}
	fmt.Println("Updated entry:", entry.Name)
	EntryTable(entry)
	return nil
}

// cmdDelete deletes an existing entry, identified by name.
func cmdDelete(c *cli.Context) error {
	name := c.String("name")
	ask := !c.Bool("yes")
	deleteEntry(name, ask)
	return nil
}

// cmdList lists entries, optionally filtered and sorted.
func cmdList(c *cli.Context) error {
	keywords := c.String("search")
	anyTags := []string{}
	if c.IsSet("tags") {
		anyTags = strings.Split(c.String("any-tags"), ",")
	}
	onlyTags := []string{}
	if c.IsSet("tag") {
		onlyTags = strings.Split(c.String("tag"), ",")
	}
	// defaults to most recent first
	order := search.SortRecent
	// unless -search is provided, then default to score
	if !c.IsSet("order") && c.IsSet("search") {
		order = search.SortScore
	}
	// or override defaults with -order
	if c.IsSet("order") {
		switch c.String("order") {
		case "name":
			order = search.SortName
		case "score":
			order = search.SortScore
		case "recent":
			order = search.SortRecent
		}
	}

	types := c.String("types")
	if interactive {
		pageSize := ListPageSize()
		results, err := memApp.Search.SearchEntries(parseTypes(types), keywords, onlyTags, anyTags,
			order, 1, pageSize)
		if err != nil {
			return err
		}
		pager := NewEntryPager(results)
		pager.PrintPage()
		if len(results.Entries) == 0 {
			return nil
		}
		if err := listInteractiveLoop(pager); err != nil {
			return err
		}
	} else {
		pageSize := util.MaxInt32
		results, err := memApp.Search.SearchEntries(parseTypes(types), keywords, onlyTags, anyTags,
			order, 1, pageSize)
		if err != nil {
			return err
		}
		EntryTables(results.Entries)
	}
	return nil
}

// cmdLinks lists the entries linked to and from an existing entry, identified by name.
func cmdLinks(c *cli.Context) error {
	name := c.String("name")
	entry, err := memApp.GetEntry(util.GetSlug(name))
	if err != nil {
		return err
	}
	if interactive {
		linksInteractiveLoop(entry)
	} else {
		LinksMenu(entry)
		fmt.Println("")
	}
	return nil
}

// cmdSeeds lists links to entries that don't exist yet
func cmdSeeds(c *cli.Context) error {
	brokenLinks, err := memApp.Search.BrokenLinks()
	if err != nil {
		return err
	}
	for from, tos := range brokenLinks {
		fmt.Println("From:", from)
		for _, to := range tos {
			fmt.Println("  ", to)
		}
	}
	return nil
}

// cmdGet displays the editable content of an entry
func cmdGet(c *cli.Context) error {
	name := c.String("name")
	entry, err := memApp.GetEntry(util.GetSlug(name))
	if err != nil {
		return err
	}
	content, err := template.RenderYamlDown(entry)
	if err != nil {
		return err
	}
	fmt.Println(content)
	return nil
}

// cmdDetail displays details of an entry and, if interactive, provides a menu prompt.
func cmdDetail(c *cli.Context) error {
	name := c.String("name")
	entry, err := memApp.GetEntry(util.GetSlug(name))
	if err != nil {
		return fmt.Errorf("entry named '%s' does not exist", name)
	} else if interactive {
		detailInteractiveLoop(entry)
	} else {
		EntryTable(entry)
	}
	return nil
}

// cmdTags displays a list of tags in use and how many entries each has
func cmdTags(c *cli.Context) error {
	tags, err := memApp.GetTags()
	if err != nil {
		return err
	}
	sorted := memApp.GetSortedTags(tags)
	fmt.Println()
	for _, tag := range sorted {
		names := tags[tag]
		fmt.Printf("%s [%d]  ", tag, len(names))
	}
	fmt.Println()
	fmt.Println()
	return nil
}

// cmdRebuild clears out the bleve index and rebuilds it from source entry files.
func cmdRebuild(c *cli.Context) error {
	return memApp.Search.Rebuild()
}

// cmdTimeline displays a timeline of entries based on start and end attributes.
func cmdTimeline(c *cli.Context) error {
	start := c.String("from")
	end := c.String("to")
	tl, err := memApp.Search.Timeline(start, end)
	if err != nil {
		return err
	}
	for _, entry := range tl {
		fmt.Println(util.Pad(entry.Start, 10, " ", false), "-",
			util.Pad(entry.End, 10, " ", false), "\t", entry.Name)
	}
	return nil
}

// cmdFiles lists files associated with an entry
func cmdFiles(c *cli.Context) error {
	entryName := c.String("entry")
	entry, err := memApp.GetEntry(util.GetSlug(entryName))
	if err != nil {
		return err
	}
	AttachmentsTable(entry.Attachments)
	return nil
}

// cmdFileAdd adds a file to an entry
func cmdFileAdd(c *cli.Context) error {
	// get arguments
	entryName := c.String("entry")
	path := c.String("path")
	name := c.String("title")
	if name == "" {
		name = util.StripExtension(path)
	}
	// get entry
	slug := util.GetSlug(entryName)
	entry, err := memApp.GetEntry(slug)
	if err != nil {
		return err
	}
	// add file
	attachment, err := memApp.Attach.Add(slug, path, name)
	if err != nil {
		return err
	}
	// attach to entry and save
	entry.Attachments = append(entry.Attachments, attachment)
	err = memApp.PutEntry(entry)
	if err != nil {
		return err
	}
	fmt.Println("File attached successfully.")
	return nil
}

// cmdFileDelete deletes a file attachment
func cmdFileDelete(c *cli.Context) error {
	entryName := c.String("entry")
	title := c.String("title")
	slug := util.GetSlug(entryName)
	entry, err := memApp.GetEntry(slug)
	if err != nil {
		return err
	}
	for ix, att := range entry.Attachments {
		if att.Name == title {
			// remove from entry
			atts := entry.Attachments
			copy(atts[ix:], atts[ix+1:])           // Shift a[i+1:] left one index.
			atts[len(atts)-1] = model.Attachment{} // Erase last element (write zero value).
			atts = atts[:len(atts)-1]              // Truncate slice.
			entry.Attachments = atts
			memApp.PutEntry(entry)
			// delete attachment
			return memApp.Attach.Delete(att)
		}
	}
	return model.FileNotFound{Path: title}
}

// cmdFileRename renames a file attachment
func cmdFileRename(c *cli.Context) error {
	entryName := c.String("entry")
	slug := util.GetSlug(entryName)
	title := c.String("title")
	newTitle := c.String("new-title")
	entry, err := memApp.GetEntry(slug)
	if err != nil {
		return err
	}
	for ix, att := range entry.Attachments {
		if att.Name == title {
			renamed, err := memApp.Attach.Rename(att, newTitle)
			if err != nil {
				return err
			}
			entry.Attachments[ix] = renamed
			memApp.PutEntry(entry)
			fmt.Println("Renamed attachment to " + renamed.Name + " (" + renamed.DisplayFileName() + ")")
			return nil
		}
	}
	return model.FileNotFound{Path: title}
}

// cmdFileOpen opens a file on the local system
func cmdFileOpen(c *cli.Context) error {
	entryName := c.String("entry")
	slug := util.GetSlug(entryName)
	title := c.String("title")
	command := c.String("command")
	if command == "" {
		command = config.OpenFileCommand
	}
	entry, err := memApp.GetEntry(slug)
	if err != nil {
		return err
	}
	for _, att := range entry.Attachments {
		if att.Name == title {
			path, err := memApp.Attach.GetAttachmentPath(att)
			if err != nil {
				return err
			}
			cmd := exec.Command(command, path)
			return cmd.Start()
		}
	}
	return model.FileNotFound{Path: title}
}
