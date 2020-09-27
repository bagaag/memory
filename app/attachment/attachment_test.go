/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package attachment

import (
	"fmt"
	"io/ioutil"
	"memory/app/model"
	"memory/util"
	"os"
	"testing"
)

// setup creates a temp folder for attachment storage and returns an instance of LocalAttachmentStore.
func setup() (LocalAttachmentStore, func() error, error) {
	tmpDir, err := ioutil.TempDir("", "attachment_test_*")
	if err != nil {
		return LocalAttachmentStore{}, nil, err
	}
	teardown := func() error {
		return util.DelTree(tmpDir)
	}
	return LocalAttachmentStore{tmpDir}, teardown, nil
}

func TestGetAttachmentPath(t *testing.T) {
	// setup and teardown
	var atts LocalAttachmentStore
	if store, teardown, err := setup(); err != nil {
		t.Error(err)
		return
	} else {
		atts = store
		defer teardown()
	}
	// test with non-existant file
	att := model.Attachment{Name: "Test Name", Extension: "txt", EntrySlug: "entry-slug"}
	path, err := atts.GetAttachmentPath(att)
	if err == nil {
		t.Error("expected FileNotFound error, got nil")
	} else if !model.IsFileNotFound(err) {
		t.Error(err)
		return
	}
	// create file
	_, err = os.Create(path)
	if err != nil {
		fmt.Println(path)
		t.Error(err)
		return
	}
	// test with existing file
	path, err = atts.GetAttachmentPath(att)
	if err != nil {
		t.Error("expected FileNotFound error, got nil")
		return
	}
}

// createTestFile creates a test file and returns the full path.
func createTestFile(contents string) (string, error) {
	file, err := ioutil.TempFile("", "test-*.txt")
	if err != nil {
		return "", err
	}
	_, err = os.Create(file.Name())
	if err != nil {
		return file.Name(), err
	}
	err = ioutil.WriteFile(file.Name(), []byte(contents), 0644)
	if err != nil {
		return file.Name(), err
	}
	return file.Name(), nil
}

// readFile returns the contents of a file or error text.
func readFile(path string) string {
	if bytes, err := ioutil.ReadFile(path); err != nil {
		return err.Error()
	} else {
		return string(bytes)
	}
}

func TestCRUD(t *testing.T) {
	// setup and teardown
	slug := "entry-slug"
	var atts LocalAttachmentStore
	if store, teardown, err := setup(); err != nil {
		t.Error(err)
		return
	} else {
		atts = store
		defer teardown()
	}
	// create file
	path, err := createTestFile("test")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(path)
	// test Add
	att, err := atts.Add(slug, path, "Test Attachment")
	if err != nil {
		t.Error(err)
		return
	}
	if att.Extension != "txt" {
		t.Error("Expected txt, got", att.Extension)
	}
	if att.Name != "Test Attachment" {
		t.Error("Expected 'Test Attachment', got", att.Name)
	}
	if att.EntrySlug != slug {
		t.Error("Expected 'entry-slug', got", att.EntrySlug)
	}
	if s := att.ExtensionWithPeriod(); s != ".txt" {
		t.Error("Expected '.txt', got", s)
	}
	if s := att.DisplayFileName(); s != "test-attachment.txt" {
		t.Error("Expected 'test-attachment.txt', got", s)
	}
	attPath, err := atts.GetAttachmentPath(att)
	if err != nil {
		t.Error(err)
		return
	}
	if s := readFile(attPath); s != "test" {
		t.Error("Expected 'test', got", s)
	}
	// test Update
	path2, err := createTestFile("test 2")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = atts.Update(slug, path2, "Not Exists")
	if !model.IsFileNotFound(err) {
		t.Error("expected FileNotFound, got", err)
	}
	att2, err := atts.Update(slug, path2, "Test Attachment")
	attPath2, err := atts.GetAttachmentPath(att2)
	if s := readFile(attPath2); s != "test 2" {
		t.Error("expected 'test 2', got", s)
	}
	// test Rename
	//Rename(entrySlug string, fileName string, newName string) (model.Attachment, error)
	// test Delete
	//Delete(entrySlug string, fileName string) error
}
