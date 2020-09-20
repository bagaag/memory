/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package model

import (
	"fmt"
	"strings"
)

// Handles metadata for entry file attachments

type Attachment struct {
	Name string
	Path string
}

// Extension returns the file extension without a leading period,
// or empty string if the file name doesn't contain a period.
func (f *Attachment) Extension() string {
	if strings.Contains(f.Path, ".") {
		parts := strings.Split(f.Path, ".")
		return parts[len(parts)-1]
	}
	return ""
}

// FileNotFound is a custom error type to indicate that a requested entry is not found in storage.
type FileNotFound struct {
	Path string
}

// IsFileNotFound returns true if err is an FileNotFound error.
func IsFileNotFound(err error) bool {
	if err != nil {
		if _, notFound := err.(FileNotFound); notFound {
			return true
		}
	}
	return false
}

// Error implements the error interface.
func (e FileNotFound) Error() string {
	return fmt.Sprintf("file %s not found", e.Path)
}
