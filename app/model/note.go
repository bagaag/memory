/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/* 
Notes are a temporary holding place where ideas that
require further development can be quickly captured.
*/

package model

import "time"

type Note struct {
  Id string
  Description string
  Tags []string
  Created time.Time
  Modified time.Time
}

