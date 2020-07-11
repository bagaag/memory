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
	"memory/app/persist"
	"os"
	"os/exec"
	"strings"
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
