/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

//Package app contains an API for interacting with the application
//that is not bound to a particular UI.
package app

import (
	"memory/app/config"
	"memory/app/model"
	"memory/app/persist"
)

type root struct {
	Notes map[string]model.Note
	//Tags  map[string]model.Tag
}

// The data variable stores all the things that get saved.
var data = root{Notes: make(map[string]model.Note)}

// Init reads data stored on the file system
// and initializes application variable.
func Init() error {

	if persist.PathExists(config.SavePath()) {
		if err := persist.Load(config.SavePath(), &data); err != nil {
			return err
		}
	}

	return nil
}

func Save() error {
	return persist.Save(config.SavePath(), data)
}
