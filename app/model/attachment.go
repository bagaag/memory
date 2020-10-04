/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package model

import (
	"fmt"
	"memory/util"
)

// Attachment handles metadata for entry file attachments.
type Attachment struct {
	// Name is the friendly/display name of the attachment.
	Name string
	// Extension is the file extension of the attachment (without period)
	Extension string
}

// ExtensionWithPeriod returns the extension with a period, or empty string if there is no extension.
func (a *Attachment) ExtensionWithPeriod() string {
	if len(a.Extension) == 0 {
		return ""
	}
	return "." + a.Extension
}

// DisplayFileName returns a display file name representing the attachment.
func (a *Attachment) DisplayFileName() string {
	return util.GetSlug(a.Name) + a.ExtensionWithPeriod()
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
