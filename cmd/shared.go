/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/*
This file contains variables and functions that are shared by the
various command-specific files comprising this package.
*/

package cmd

import (
	"fmt"
	"memory/app"
	"memory/app/config"
	"memory/app/model"
	"memory/app/persist"
	"os"
	"os/exec"
	"strings"

	"github.com/buger/goterm"
	"github.com/olekukonko/tablewriter"
)

// processTags takes in a comma-separated string and returns a slice of trimmed values
func processTags(tags string) []string {
	if strings.TrimSpace(tags) == "" {
		return []string{}
	}
	arr := strings.Split(tags, ",")
	for i, tag := range arr {
		arr[i] = strings.ToLower(strings.TrimSpace(tag))
	}
	return arr
}

// Shared function for use by commands to save data to config.SavePath.
func save() {
	if err := app.Save(); err != nil {
		fmt.Println("Failed to save data:", err)
	}
}

// Displays a table of entries
func displayEntries(entries []model.Entry, full bool) {
	width := goterm.Width() - 30
	fmt.Println("") // prefix with blank line
	for ix, entry := range entries {
		switch entry.(type) {
		case model.Note:
			// holds table contents
			data := [][]string{}
			// add note name row
			note := entry.(model.Note)
			data = append(data, []string{"Name", note.Name()})
			// description row
			desc := note.Description()
			if !full && len(desc) > config.TruncateAt {
				desc = desc[:config.TruncateAt] + "..."
			}
			if desc != "" {
				data = append(data, []string{"Description", desc})
			}
			// tags row
			if len(note.Tags()) > 0 {
				data = append(data, []string{"Tags", strings.Join(note.Tags(), ", ")})
			}
			// create and configure table
			table := tablewriter.NewWriter(os.Stdout)
			// add border to top unless this is the first
			if ix > 0 {
				table.SetBorders(tablewriter.Border{Left: false, Top: true, Right: false, Bottom: false})
			} else {
				table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
			}
			table.SetRowLine(false)
			table.SetColMinWidth(0, 12)
			table.SetColMinWidth(1, width)
			table.SetColWidth(width)
			table.SetAutoWrapText(true)
			table.SetReflowDuringAutoWrap(true)
			// add data and render
			table.AppendBulk(data)
			table.Render()
		}
	}
	fmt.Println("") // finish with blank line
}

// subPrompt asks for additional info within a command.
func subPrompt(prompt string, validate validator) string {
	rl.HistoryDisable()
	rl.SetPrompt(prompt)
	var err error
	var input = ""
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
	return strings.TrimSpace(input)
}

// subPromptEditor provides an option to use a full editor for
// a long-text value rather than readline.
func subPromptEditor(fieldName string, value string, prompt string, validate validator) string {
	useEditor := subPrompt(fieldName+" may hold paragraphs of text. Would you like to use a full screen editor? [Y/n]: ", validateYesNo)
	useEditor = strings.ToLower(strings.TrimSpace(useEditor))
	if useEditor == "y" || useEditor == "" {
		var tmp string
		var err error
		if tmp, err = persist.CreateTempFile(value); err != nil {
			fmt.Println("Failed to create temporary file:", err)
			return subPrompt(prompt, validate)
		}
		cmd := exec.Command(config.EditorCommand, tmp)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			fmt.Println("Failed to interact with editor:", err)
			return ""
		}
		var edited string
		if edited, err = persist.ReadTempFile(tmp); err != nil {
			fmt.Println("Failed to read temporary file:", err)
			return ""
		}
		fmt.Println("Retrieved content from temporary file.")
		if err := persist.RemoveTempFile(tmp); err != nil {
			fmt.Println("Failed to delete temporary file:", err)
		}
		return edited
	}
	// user elected not to use editor; use std readline prompt
	return subPrompt(prompt, validate)
}
