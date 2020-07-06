/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

// The persist package handles persistence tasks.

package persist

import (
	"bytes"
	"encoding/json"
  "errors"
	"io"
	"os"
	"sync"
)

var lock sync.Mutex

// Marshals the object into an io.Reader
func Marshal(v interface{}) (io.Reader, error) {
  b, err := json.MarshalIndent(v, "", "\t")
  if err != nil {
    return nil, err
  }
  return bytes.NewReader(b), nil
}

// Save saves a representation of v to the file at path
func Save(path string, v interface{}) error {
  lock.Lock()
  defer lock.Unlock()
  f, err := os.Create(path)
  if err != nil {
    return err
  }
  defer f.Close()
  r, err := Marshal(v)
  if err != nil {
    return err
  }
  _, err = io.Copy(f, r)
  return err
}

// Unmarshals data from the reader into the specified value
func Unmarshal(r io.Reader, v interface{}) error {
  return json.NewDecoder(r).Decode(v)
}

// Load the json file at path into v
func Load(path string, v interface{}) error {
  lock.Lock()
  defer lock.Unlock()
  f, err := os.Open(path)
  if err != nil {
    return err
  }
  defer f.Close()
  return Unmarshal(f, v)
}

// Returns true if the given path exists.
func PathExists(path string) bool {
  if _, err := os.Stat(path); err != nil {
    if errors.Is(err, os.ErrNotExist) {
      return false
    }
  }
  return true
}
