/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package model

import (
	"fmt"
)

// Handles metadata for entry file attachments

type Attachment struct {
	Name string
	Path string
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
