/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/* This file contains tests for the functions in util.go. */

package util

import "testing"

func TestStringSlicesEqual(t *testing.T) {
	ss1 := []string{"One", "TWO", "three"}
	ss2 := []string{"One", "TWO", "three"}
	ss3 := []string{"One", "TWO", "four"}
	ss4 := []string{"One", "three"}
	if !StringSlicesEqual(ss1, ss2) {
		t.Errorf("%s should equal %s", ss1, ss2)
	}
	if StringSlicesEqual(ss2, ss3) {
		t.Errorf("%s should not equal %s", ss2, ss3)
	}
	if StringSlicesEqual(ss1, ss4) {
		t.Errorf("%s should not equal %s", ss1, ss4)
	}
}

func TestStringSliceToLower(t *testing.T) {
	ss := []string{"One", "TWO", "three"}
	expect := []string{"one", "two", "three"}
	StringSliceToLower(ss)
	if !StringSlicesEqual(ss, expect) {
		t.Errorf("Expected %s, got %s", expect, ss)
	}
}

func TestStringSliceContains(t *testing.T) {
	ss := []string{"one", "Two", "three"}
	if !StringSliceContains(ss, "Two") {
		t.Errorf("%s should contain %s", ss, "Two")
	}
	if StringSliceContains(ss, "two") {
		t.Errorf("%s should not contain %s", ss, "two")
	}
	if StringSliceContains(ss, "four") {
		t.Errorf("%s should not contain %s", ss, "four")
	}
}
