/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

/* This file contains tests for the functions in util.go. */

package util

import (
	"fmt"
	"testing"
)

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

func TestTruncateAtWhitespace(t *testing.T) {
	sa := "One  two three\nFour five \t six seven."
	sb := "One  two three\nFour five \t sixes seven."
	s1 := TruncateAtWhitespace(sa, 28)
	expect1 := "One two three Four five six"
	if len(s1) > 28 {
		t.Error("1. Expected len <= 28, got", len(s1))
	}
	if s1 != expect1 {
		t.Error("1. Expected", expect1, "got", s1)
	}
	s2 := TruncateAtWhitespace(sb, 28)
	expect2 := "One two three Four five"
	if len(s2) > 28 {
		t.Error("2. Expected len <= 28, got", len(s2))
	}
	if s2 != expect2 {
		t.Error("2. Expected", expect2, "got", s2)
	}
	s3 := TruncateAtWhitespace("", 28)
	if s3 != "" {
		t.Error("3. Expected '', got ", s3)
	}
}

func TestPad(t *testing.T) {
	s := "x"
	left := Pad(s, 3, " ", true)
	right := Pad(s, 3, " ", false)
	if left != "  x" {
		fmt.Errorf("Expected '  x' got ''%s", left)
	}
	if right != "x  " {
		fmt.Errorf("Expected 'x  ' got ''%s", right)
	}
}
