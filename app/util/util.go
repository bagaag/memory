/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package util

import "strings"

// FormatErrorForDisplay takes an error message, which idiomatically should not be capitalized or
// in sentence format, and returns a string with the first letter capitalized and a period at the
// end.
func FormatErrorForDisplay(err error) string {
	var s string
	s = err.Error()
	return strings.ToUpper(s[:1]) + s[1:] + "."
}
