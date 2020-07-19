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
	"memory/app/config"
	"memory/app/persist"
	"os"
	"os/exec"
	"strconv"
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

// subPrompt asks for additional info within a command.
func subPrompt(prompt string, value string, validate validator) (string, error) {
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

// subPromptEditor provides an option to use a full editor for
// a long-text value rather than readline.
func subPromptEditor(fieldName string, value string, prompt string, validate validator) (string, error) {
	useEditor, err := subPrompt(fieldName+" may hold paragraphs of text. Would you like to use a full screen editor? [Y/n]: ", "", validateYesNo)
	if err != nil {
		return "", err
	}
	useEditor = strings.ToLower(strings.TrimSpace(useEditor))
	if useEditor == "y" || useEditor == "" {
		var tmp string
		if tmp, err = persist.CreateTempFile(value); err != nil {
			fmt.Println("Failed to create temporary file:", err)
			return subPrompt(prompt, value, validate)
		}
		cmd := exec.Command(config.EditorCommand, tmp)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			fmt.Println("Failed to interact with editor:", err)
			return "", err
		}
		var edited string
		if edited, err = persist.ReadTempFile(tmp); err != nil {
			fmt.Println("Failed to read temporary file:", err)
			return "", err
		}
		fmt.Println("Retrieved content from temporary file.")
		if err := persist.RemoveTempFile(tmp); err != nil {
			fmt.Println("Failed to delete temporary file:", err)
		}
		return edited, nil
	}
	// user elected not to use editor; use std readline prompt
	return subPrompt(prompt, value, validate)
}

// listPrompt presents a numbered list of options and prompts the user to choose one.
// TODO: Handle interrupts for all prompt functions
func listPrompt(prompt string, list []string) (int, error) {
	rl.HistoryDisable()
	fmt.Println(prompt)
	for i, v := range list {
		fmt.Printf("%3d. %s\n", i+1, v)
	}
	rl.SetPrompt("[1]" + config.SubPrompt)
	var err error
	var input string
	selection := 1
	for {
		input, err = rl.Readline()
		if err != nil {
			break
		}
		if input == "" {
			return 1, nil
		}
		if i, err := strconv.Atoi(input); err != nil || i <= 0 || i > len(list) {
			fmt.Println("Enter a number from 1 to", len(list))
		} else {
			selection = i - 1
			break
		}
	}
	rl.HistoryEnable()
	rl.SetPrompt(config.Prompt)
	return selection, err
}
