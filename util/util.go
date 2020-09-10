/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package util

import (
	"fmt"
	"github.com/gosimple/slug"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/term"
)

// MaxInt32 is the max value that can be assigned to int32
var MaxInt32 = 2147483647

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

// StringSliceToLower converts all the strings in a slice to lower case.
func StringSliceToLower(ss []string) {
	for i, s := range ss {
		ss[i] = strings.ToLower(s)
	}
}

// Indent the text, preserving line breaks.
func Indent(s string, n int) string {
	lines := strings.Split(s, "\n")
	for ix, line := range lines {
		lines[ix] = "  " + line
	}
	return strings.Join(lines, "\n")
}

// DelTree deletes a folder and all of its contents
func DelTree(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return os.Remove(dir)
}

// GetSlug converts a string into a slug
func GetSlug(s string) string {
	return slug.Make(s)
}

// TruncateAtWhitespace returns a subset of the given string with a length equal to or less than
// the given length at a whitespace breakpoint.
func TruncateAtWhitespace(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	for strings.Index(s, "  ") != -1 {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	if len(s) <= maxLen {
		return s
	}
	words := strings.Split(s, " ")
	ix := 0
	length := 0
	for {
		length = length + len(words[ix]) + 1
		if length > maxLen {
			break
		}
		ix = ix + 1
	}
	return strings.Join(words[:ix], " ")
}

// Max date allowed in bleve queries
func MaxRFC3339Time() time.Time {
	d, _ := time.Parse(time.RFC3339, "2262-04-11T11:59:59Z")
	return d
}

// Min date allowed in bleve queries
func MinRFC3339Time() time.Time {
	d, _ := time.Parse(time.RFC3339, "1677-12-01T00:00:00Z")
	return d
}

// Pad adds the string to the end or beginning of the string until the specified length is reached
func Pad(s string, length int, padding string, before bool) string {
	for len(s) < length {
		if before {
			s = padding + s
		} else {
			s = s + padding
		}
	}
	return s
}
