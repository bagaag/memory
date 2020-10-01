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
	"memory/util"
)

// Attacher is an interface for managing entry attachments.
type Attacher interface {
	// GetAttachmentPath returns the complete file system path for an attachment for viewing or editing.
	GetAttachmentPath(attachment model.Attachment) (string, error)
	// Add returns a file object after copying a local file path into the attachment store.
	Add(entrySlug string, physicalPath string, friendlyName string) (model.Attachment, error)
	// Update commits a modified attachment file to the attachment store.
	Update(attachment model.Attachment, physicalPath string) (model.Attachment, error)
	// Delete removes an attachment from the store.
	Delete(attachment model.Attachment) error
	// Rename updates an attachment to reflect a new friendly name and returns an updated File object.
	Rename(attachment model.Attachment, newName string) (model.Attachment, error)
}

// LocalAttachmentStore implements the Attacher interface using local file storage.
type LocalAttachmentStore struct {
	// StoragePath is the file system location where attachments will be stored, should not end with a slash.
	StoragePath string
}

// resolvePath returns the file system path for an attachment.
func (a *LocalAttachmentStore) resolvePath(attachment model.Attachment) string {
	return a.StoragePath + localfs.Slash + attachment.EntrySlug + "-" + util.GetSlug(attachment.Name) + attachment.ExtensionWithPeriod()
}

// GetAttachmentPath returns the complete file system path for an attachment for viewing or editing.
func (a *LocalAttachmentStore) GetAttachmentPath(attachment model.Attachment) (string, error) {
	path := a.resolvePath(attachment)
	if !localfs.PathExists(path) {
		return path, model.FileNotFound{Path: path}
	}
	return path, nil
}

// Add returns a file object after copying a local file path into the attachment store.
func (a *LocalAttachmentStore) Add(entrySlug string, physicalPath string, friendlyName string) (model.Attachment, error) {
	attachment := model.Attachment{EntrySlug: entrySlug, Name: friendlyName, Extension: util.Extension(physicalPath)}
	path := a.resolvePath(attachment)
	if localfs.PathExists(path) {
		return attachment, errors.New("an attachment with this name already exists")
	}
	if err := localfs.CopyFile(physicalPath, path); err != nil {
		return attachment, err
	}
	return attachment, nil
}

// Update commits a modified attachment file to the attachment store.
func (a *LocalAttachmentStore) Update(attachment model.Attachment, physicalPath string) (model.Attachment, error) {
	path := a.resolvePath(attachment)
	if !localfs.PathExists(path) {
		return attachment, model.FileNotFound{Path: path}
	}
	if err := localfs.RemoveFile(path); err != nil {
		return attachment, err
	}
	if err := localfs.CopyFile(physicalPath, path); err != nil {
		return attachment, err
	}
	return attachment, nil
}

// Delete removes an attachment from the store.
func (a *LocalAttachmentStore) Delete(attachment model.Attachment) error {
	path := a.resolvePath(attachment)
	if !localfs.PathExists(path) {
		return model.FileNotFound{Path: path}
	}
	return localfs.RemoveFile(path)
}

// Rename updates an attachment to reflect a new friendly name and returns an updated File object.
func (a *LocalAttachmentStore) Rename(attachment model.Attachment, newName string) (model.Attachment, error) {
	oldPath := a.resolvePath(attachment)
	newAttachment := model.Attachment{EntrySlug: attachment.EntrySlug, Extension: attachment.Extension, Name: newName}
	newPath := a.resolvePath(newAttachment)
	if !localfs.PathExists(oldPath) {
		return attachment, model.FileNotFound{Path: oldPath}
	}
	if localfs.PathExists(newPath) {
		return newAttachment, errors.New("attachment with this name already exists")
	}
	if err := localfs.CopyFile(oldPath, newPath); err != nil {
		return attachment, err
	}
	if err := localfs.RemoveFile(oldPath); err != nil {
		return newAttachment, err
	}
	return newAttachment, nil
}
