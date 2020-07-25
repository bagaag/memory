/*
This file is part of the software application Memory
See https://github.com/bagaag/memory
Copyright © 2020 Matt Wiseley
License: https://www.gnu.org/licenses/gpl-3.0.txt
*/

package link

import (
	"memory/app"
	"memory/app/model"
	"memory/app/util"
	"testing"
)

func TestParseLinks(t *testing.T) {
	app.PutEntry(model.NewNote("Exists", "", []string{}))
	app.PutEntry(model.NewNote("Exists 2", "", []string{}))
	testParseLinks(t, 1, "[Exists]", "[Exists]", []string{"Exists"})
	testParseLinks(t, 2, "text [Exists]", "text [Exists]", []string{"Exists"})
	testParseLinks(t, 3, "[Exists] text", "[Exists] text", []string{"Exists"})
	testParseLinks(t, 4, "[Not Exists]", "[!Not Exists]", []string{})
	testParseLinks(t, 5, "[Exists] [Exists  \n2]", "[Exists] [Exists  \n2]", []string{"Exists", "Exists 2"})
	testParseLinks(t, 6, "", "", []string{})
	testParseLinks(t, 7, "[Exists]\n[Exists 2]", "[Exists]\n[Exists 2]", []string{"Exists", "Exists 2"})
	testParseLinks(t, 8, "[~Exists]", "[~Exists]", []string{})
	testParseLinks(t, 9, "[!Exists]", "[Exists]", []string{"Exists"})
	testParseLinks(t, 10, "[!Not Exists]", "[!Not Exists]", []string{})
	testParseLinks(t, 11, "[~Not Exists]", "[~Not Exists]", []string{})
	testParseLinks(t, 12, "[Exists 2]\n[Exists]", "[Exists 2]\n[Exists]", []string{"Exists 2", "Exists"})
}

func testParseLinks(t *testing.T, testNo int, input string, parsedExpected string, linksExpected []string) {
	parsed, links := ParseLinks(input)
	if parsed != parsedExpected {
		t.Errorf("#%d Expected parsed '%s', got '%s'", testNo, parsedExpected, parsed)
	}
	if !util.StringSlicesEqual(linksExpected, links) {
		t.Errorf("#%d Expected links '%s', got '%s'", testNo, linksExpected, links)
	}
}

func TestResolveLinks(t *testing.T) {
	app.PutEntry(model.NewNote("Exists", "", []string{}))
	app.PutEntry(model.NewNote("Exists 2", "", []string{}))
	links := []string{"Exists", "Not exists", "Exists 2"}
	resolved := ResolveLinks(links)
	if len(resolved) != 2 {
		t.Error("Expected len of 2, got", len(resolved))
	}
	if resolved[1].Name() != "Exists 2" {
		t.Error("Expected 'Exists 2', got", resolved[1].Name())
	}
}
