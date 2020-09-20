/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package attachment

import (
	"errors"
	"memory/app/model"
)

// Attacher is an interface for managing entry attachments.
type Attacher interface {
	// Path returns the complete file system path for an attachment.
	Path(entrySlug string, filePath string) (string, error)
	// Add returns a file object after copying a local file path into the attachment store.
	Add(entrySlug string, physicalPath string, friendlyName string) (model.File, error)
	// Update commits a modified attachment file to the attachment store.
	Update(entrySlug string, physicalPath string, friendlyName string) (model.File, error)
	// Delete removes an attachment from the store.
	Delete(entrySlug string, fileName string) error
	// Rename updates an attachment to reflect a new friendly name and returns an updated File object.
	Rename(entrySlug string, fileName string, newName string) (model.File, error)
}

// LocalFileStore implements the Attacher interface using local file storage.
type LocalFileStore struct {
}

// Path returns the complete file system path for an attachment.
func (a *LocalFileStore) Path(entrySlug string, filePath string) (string, error) {
	return "", errors.New("not implemented")
}

// Add returns a file object after copying a local file path into the attachment store.
func (a *LocalFileStore) Add(entrySlug string, physicalPath string, friendlyName string) (model.File, error) {
	return model.File{}, errors.New("not implemented")
}

// Update commits a modified attachment file to the attachment store.
func (a *LocalFileStore) Update(entrySlug string, physicalPath string, friendlyName string) (model.File, error) {
	return model.File{}, errors.New("not implemented")
}

// Delete removes an attachment from the store.
func (a *LocalFileStore) Delete(entrySlug string, fileName string) error {
	return errors.New("not implemented")
}

// Rename updates an attachment to reflect a new friendly name and returns an updated File object.
func (a *LocalFileStore) Rename(entrySlug string, fileName string, newName string) (model.File, error) {
	return model.File{}, errors.New("not implemented")
}
