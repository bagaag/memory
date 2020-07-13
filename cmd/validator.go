/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/*
This file contains validators used by sub-prompts in interactive mode.
*/

package cmd

import (
	"fmt"
	"memory/app"
	"memory/app/config"
	"strings"
)

// validator is a function that validates input and returns either an error
// message (failure) or empty string (success)
type validator func(input string) string

// validateName checks for general name issues, regardless of type
func validateName(name string) string {
	if name == "" {
		return "A unique name is required."
	}
	if len(name) > config.MaxNameLen {
		return fmt.Sprintf("Names must be 50 or fewer characters. This one is %d characters.", len(name))
	}
	return ""
}

// validateNoteName checks for general name issues and that the name isn't already in use
func validateNoteName(name string) string {
	if msg := validateName(name); msg != "" {
		return msg
	}
	if _, exists := app.GetEntry(name); exists {
		return "A unique name is required. This name is already in use."
	}
	return ""
}

func emptyValidator(s string) string {
	return ""
}

func validateYesNo(answer string) string {
	answer = strings.ToLower(strings.TrimSpace(answer))
	if answer == "y" || answer == "n" || answer == "" {
		return ""
	}
	return "Respond with y, n or nothing at all to accept the default."
}
