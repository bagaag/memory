/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package model

import "time"

// Entry contains accessors for common model fields (people, places, things, events)
type Entry interface {
	Name() string
	Description() string
	Tags() []string
	Created() time.Time
	Modified() time.Time
}
