/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package persist

import (
	"io/ioutil"
	"memory/app/config"
	"memory/app/localfs"
	"memory/util"
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
	if err := localfs.Save(path, v); err != nil {
		t.Errorf("Error saving file: %s", err)
		return
	}
	var v2 testStruct
	if err := localfs.Load(path, &v2); err != nil {
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
	tempDir, err := ioutil.TempDir("", "test_temp_file")
	if err != nil {
		t.Error(err)
		return
	}
	config.MemoryHome = tempDir
	os.Mkdir(tempDir+string(os.PathSeparator)+"tmp", 0740)
	defer util.DelTree(tempDir)
	temp := "one\n\two\three"
	if path, err := localfs.CreateTempFile("test-temp-file", temp); err != nil {
		t.Errorf("%s", err)
	} else {
		if s, _, err2 := localfs.ReadFile(path); err2 != nil {
			t.Errorf("%s", err)
		} else if s != temp {
			t.Errorf("%s != %s", temp, s)
		}
		localfs.RemoveFile(path)
	}
}
