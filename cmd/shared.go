/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/* This file contains functions that are used throughout the cmd package. */

package cmd

import (
	"errors"
	"fmt"
	"memory/app/config"
	"memory/app/links"
	"memory/app/localfs"
	"memory/app/model"
	"memory/app/template"
	"memory/util"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/chzyer/readline"
)

// filterInput allows certain keys to be intercepted during readline
func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

// parseTypes populates an EntryType struct based on the --types flag
func parseTypes(typesArg string) model.EntryTypes {
	typesSlice := strings.Split(typesArg, ",")
	types := model.EntryTypes{}
	for _, t := range typesSlice {
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

// editEntry converts an entry to YamlDown, launches an external editor, parses
// the edited content back into an entry and returns the edited entry.
func editEntry(origEntry model.Entry, tempFile string) (model.Entry, string, error) {
	var err error
	// launch editor and get path to edited temp file
	tempFile, err = useEditor(origEntry, tempFile)
	if err != nil {
		return model.Entry{}, tempFile, err
	}
	// get contents of temp file
	edited, _, err := localfs.ReadFile(tempFile)
	if err != nil {
		return model.Entry{}, tempFile, err
	}
	// parse contents into entry
	editedEntry, err := parseEntryText(edited)
	if err != nil {
		return model.Entry{}, tempFile, err
	}
	// update attachment titles
	// TODO: figure out better way than index to connect edited file back to orig
	for ix, updatedAtt := range editedEntry.Attachments {
		origAtt := origEntry.Attachments[ix]
		if origAtt.Name != updatedAtt.Name {
			updatedAtt, err = memApp.Attach.Rename(editedEntry.Slug(), origAtt, updatedAtt.Name)
			if err != nil {
				return editedEntry, tempFile, err
			}
			editedEntry.Attachments[ix] = updatedAtt
		}
	}
	// handle name change
	if origEntry.Name != editedEntry.Name {
		if memApp.EntryExists(editedEntry.Slug()) {
			return editedEntry, tempFile, errors.New("entry named '" + editedEntry.Name + "' already exists")
		}
		if memApp.EntryExists(origEntry.Slug()) {
			if err = memApp.DeleteEntry(origEntry.Slug()); err != nil {
				return editedEntry, tempFile, err
			}
			//TODO: update links on rename
		}
	}
	// save changes
	if !memApp.EntryExists(editedEntry.Slug()) {
		editedEntry.Created = time.Now()
	}
	editedEntry.Modified = time.Now()
	editedEntry.Description = links.RenderLinks(editedEntry.Description, memApp.EntryExists)
	if err = memApp.PutEntry(editedEntry); err != nil {
		return editedEntry, tempFile, err
	}
	return editedEntry, "", nil
}

// parseEntryText converts text to an entry and validates the name.
func parseEntryText(entryText string) (model.Entry, error) {
	editedEntry, err := template.ParseYamlDown(entryText)
	if err != nil {
		return model.Entry{}, err
	}
	if msg := validateName(editedEntry.Name); msg != "" {
		return editedEntry, errors.New(msg)
	}
	return editedEntry, nil
}

// deleteEntry deletes the entry, saves, and prints an error if any. Returns true if successful.
func deleteEntry(name string, ask bool) bool {
	s := "y"
	var err error
	if !memApp.EntryExists(util.GetSlug(name)) {
		fmt.Println("Entry '" + name + "' could not be found.")
		return false
	}
	if ask {
		s, err = subPrompt("Are you sure you want to delete "+name+"? [y,N]: ", "", validateYesNo)
		if err != nil {
			fmt.Println("Error:", err)
			return false
		}
	}
	if s == "y" {
		if err := memApp.DeleteEntry(util.GetSlug(name)); err != nil {
			fmt.Println("Error:", err)
			return false
		}
		fmt.Println("Entry deleted.")
		return true
	}
	return false
}

// useEditor launches config.editor with a temporary file containing a copy of the entry
// identified by slug, waits for the editor to exit and returns the temp file path. If
// existingTempFilePath is not empty, reuses that file rather than creating a new copy.
func useEditor(entry model.Entry, existingTempFile string) (string, error) {
	tmp := existingTempFile
	var err error
	var content string
	slug := entry.Slug()
	if tmp == "" {
		editableEntry, err := memApp.GetEntry(slug)
		if err != nil {
			if _, notFound := err.(model.EntryNotFound); !notFound {
				return "", err
			}
			// entry doesn't exist
			content, err = template.RenderYamlDown(entry)
		} else {
			// entry exists
			content, err = template.RenderYamlDown(editableEntry)
		}
		if err != nil {
			return "", fmt.Errorf("failed to render new entry: %s", err.Error())
		}
		tmp, err = localfs.CreateTempFile(slug, content)
		if err != nil {
			return "", fmt.Errorf("failed to create temporary file: %s", err.Error())
		}
	}
	cmd := exec.Command(config.EditorCommand, tmp)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to interact with editor: %s", err.Error())
	}
	return tmp, nil
}

// Displays prompt for single character input and returns the character entered, or empty string.
func getSingleCharInput() string {
	fmt.Print(config.SubPrompt)
	ascii, _, err := util.ReadKeyStroke()
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

// subPrompt asks for additional info within a command.
func subPrompt(prompt string, value string, validate validator) (string, error) {
	if rl == nil {
		return "", errors.New("readline not initialized")
	}
	rl.HistoryDisable()
	rl.SetPrompt(prompt)
	var err error
	var input = value
	for {
		input, err = rl.ReadlineWithDefault(input)
		if err != nil {
			break
		}
		if msg := validate(input); msg != "" {
			fmt.Println(msg)
		} else {
			break
		}
	}
	rl.HistoryEnable()
	rl.SetPrompt(config.Prompt)
	return strings.TrimSpace(input), err
}
