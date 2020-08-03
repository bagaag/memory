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
	"github.com/pkg/term"
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

// ReadKeyStroke returns either an ascii code, or (if input is an arrow) a Javascript key code.
func ReadKeyStroke() (ascii int, keyCode int, err error) {
	t, err := term.Open("/dev/tty")
	if err != nil {
		return
	}
	err = term.RawMode(t)
	if err != nil {
		return
	}
	bytes := make([]byte, 3)

	var numRead int
	numRead, err = t.Read(bytes)
	if err != nil {
		return
	}
	if numRead == 3 && bytes[0] == 27 && bytes[1] == 91 {
		// Three-character control sequence, beginning with "ESC-[".

		// Since there are no ASCII codes for arrow keys, we use
		// Javascript key codes.
		if bytes[2] == 65 {
			// Up
			keyCode = 38
		} else if bytes[2] == 66 {
			// Down
			keyCode = 40
		} else if bytes[2] == 67 {
			// Right
			keyCode = 39
		} else if bytes[2] == 68 {
			// Left
			keyCode = 37
		}
	} else if numRead == 1 {
		ascii = int(bytes[0])
	} else {
		// Two characters read??
	}
	t.Restore()
	t.Close()
	return
}
