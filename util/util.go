/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
)

// FormatErrorForDisplay takes an error message, which idiomatically should not be capitalized or
// in sentence format, and returns a string with the first letter capitalized and a period at the
// end.
func FormatErrorForDisplay(err error) string {
	var s string
	s = err.Error()
	return strings.ToUpper(s[:1]) + s[1:] + "."
}

// StringSliceContains returns true if a slice of strings contains the given string.
func StringSliceContains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

// StringSlicesEqual returns true if the two strings slices contain the same values.
func StringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// GetHomeDir returns the path to the user's home directory, falling back to cwd and then ".".
func GetHomeDir() string {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println("Could not find home directory:", err)
		// Fail gracefully and use current working directory if home can't be located
		if home, err = os.Getwd(); err != nil {
			fmt.Println("Could not find working directory:", err)
			home = "."
		}
	}
	return home
}
