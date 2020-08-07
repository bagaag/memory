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

// ValidateName checks for general name issues, regardless of type
func validateName(name string) string {
	if name == "" {
		return "A unique name is required."
	}
	if len(name) > config.MaxNameLen {
		return fmt.Sprintf("Names must be 50 or fewer characters. This one is %d characters.", len(name))
	}
	return ""
}

func validateType(t string) string {
	if t != app.EntryTypeEvent && t != app.EntryTypePerson && t != app.EntryTypePlace &&
		t != app.EntryTypeThing && t != app.EntryTypeNote {
		return fmt.Sprintf("Type is not one of the valid entry types (%s, %s, %s, %s, %s).",
			app.EntryTypeEvent, app.EntryTypePerson, app.EntryTypePlace, app.EntryTypeThing, app.EntryTypeNote)
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
