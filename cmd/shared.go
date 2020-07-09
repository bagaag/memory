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
	"os"
	"strings"

	"github.com/buger/goterm"
	"github.com/olekukonko/tablewriter"
)

// processTags takes in a comma-separated string and returns a slice of trimmed values
func processTags(tags string) []string {
	arr := strings.Split(tags, ",")
	for i, tag := range arr {
		arr[i] = strings.ToLower(strings.TrimSpace(tag))
	}
	return arr
}

// Shared function for use by commands to save data to config.Savepath.
func save() {
	if err := app.Save(); err != nil {
		fmt.Println("Failed to save data:", err)
	}
}

// Displays a table of entries
func displayEntries(entries []model.Entry, full bool) {
	width := goterm.Width() - 30
	fmt.Println("")
	for ix, entry := range entries {
		switch entry.(type) {
		case model.Note:
			data := [][]string{}
			note := entry.(model.Note)
			data = append(data, []string{"Name", note.Name()})
			desc := note.Description()
			if !full && len(desc) > config.TruncateAt {
				desc = desc[:config.TruncateAt] + "..."
			}
			if desc != "" {
				data = append(data, []string{"Description", desc})
			}
			if len(note.Tags()) > 0 {
				data = append(data, []string{"Tags", strings.Join(note.Tags(), ", ")})
			}
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
			table.AppendBulk(data)
			table.Render()
		}
	}
	fmt.Println("")
}
