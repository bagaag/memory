/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package persist

import (
	"os"
	"testing"
	"time"
)

type testStruct struct {
	Name    string
	Child   subStruct
	Created time.Time
}

type subStruct struct {
	Name    string
	Created time.Time
}

func TestRoundTrip(t *testing.T) {
	path := "test_file"
	sub := subStruct{"child", time.Now()}
	v := testStruct{"test", sub, time.Now()}
	if err := Save(path, v); err != nil {
		t.Errorf("Error saving file: %s", err)
		return
	}
	var v2 testStruct
	if err := Load(path, &v2); err != nil {
		t.Errorf("Error saving file: %s", err)
		return
	}
	if v2.Name != "test" || v2.Child.Name != "child" {
		t.Errorf("'%s' != 'test' || '%s' != 'child'", v2.Name, v2.Child.Name)
	}
	if err := os.Remove(path); err != nil {
		t.Errorf("could not delete test file %s", err)
	}
}

func TestTempFile(t *testing.T) {
	temp := "one\n\two\three"
	if path, err := CreateTempFile(temp); err != nil {
		t.Errorf("%s", err)
	} else {
		if s, err2 := ReadFile(path); err2 != nil {
			t.Errorf("%s", err)
		} else if s != temp {
			t.Errorf("%s != %s", temp, s)
		}
	}
}
