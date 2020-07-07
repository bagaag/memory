/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/*
CRUD methods for Notes
*/

package app

import (
	"fmt"
	"memory/app/model"
	"time"
)

func NewNote(description string, tags []string) model.Note {
	id := string(fmt.Sprintf("note-%d", len(data.Notes)+1))
	now := time.Now()
	note := model.Note{
		Id:          id,
		Description: description,
		Tags:        tags,
		Created:     now,
		Modified:    now,
	}
	return note
}

func GetNote(id string) model.Note {
	return data.Notes[id]
}

func PutNote(note model.Note) {
	data.Notes[note.Id] = note
}

func DeleteNote(id string) {
}
