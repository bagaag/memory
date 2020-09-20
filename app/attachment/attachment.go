/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package attachment

import (
	"errors"
	"memory/app/localfs"
	"memory/app/model"
)

// Attacher is an interface for managing entry attachments.
type Attacher interface {
	// GetAttachment returns the complete file system path for an attachment for viewing or editing.
	GetAttachment(entrySlug string, fileName string) (string, error)
	// Add returns a file object after copying a local file path into the attachment store.
	Add(entrySlug string, physicalPath string, friendlyName string) (model.Attachment, error)
	// Update commits a modified attachment file to the attachment store.
	Update(entrySlug string, physicalPath string, friendlyName string) (model.Attachment, error)
	// Delete removes an attachment from the store.
	Delete(entrySlug string, fileName string) error
	// Rename updates an attachment to reflect a new friendly name and returns an updated File object.
	Rename(entrySlug string, fileName string, newName string) (model.Attachment, error)
}

// LocalAttachmentStore implements the Attacher interface using local file storage.
type LocalAttachmentStore struct {
	// StoragePath is the file system location where attachments will be stored, should not end with a slash.
	StoragePath string
}

// resolvePath returns the file system path for an attachment and a boolean indicating its existence..
func (a *LocalAttachmentStore) resolvePath(entrySlug string, fileName string) (string, bool) {
	path := a.StoragePath + localfs.Slash + fileName
	return path, localfs.PathExists(path)
}

// GetAttachment returns the complete file system path for an attachment for viewing or editing.
func (a *LocalAttachmentStore) GetAttachment(entrySlug string, fileName string) (string, error) {
	path, exists := a.resolvePath(entrySlug, fileName)
	if !exists {
		return "", model.FileNotFound{Path: path}
	}
	return path, nil
}

// Add returns a file object after copying a local file path into the attachment store.
func (a *LocalAttachmentStore) Add(entrySlug string, physicalPath string, friendlyName string) (model.Attachment, error) {

	path := a.resolvePath(entrySlug)
	return model.Attachment{}, errors.New("not implemented")
}

// Update commits a modified attachment file to the attachment store.
func (a *LocalAttachmentStore) Update(entrySlug string, physicalPath string, friendlyName string) (model.Attachment, error) {
	return model.Attachment{}, errors.New("not implemented")
}

// Delete removes an attachment from the store.
func (a *LocalAttachmentStore) Delete(entrySlug string, fileName string) error {
	return errors.New("not implemented")
}

// Rename updates an attachment to reflect a new friendly name and returns an updated File object.
func (a *LocalAttachmentStore) Rename(entrySlug string, fileName string, newName string) (model.Attachment, error) {
	return model.Attachment{}, errors.New("not implemented")
}
