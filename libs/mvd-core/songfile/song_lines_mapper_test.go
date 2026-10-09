package songfile

import (
	"reflect"
	"testing"
)

func TestParseLines(t *testing.T) {
	text := "\ufeff# my list\n" +
		"ATB - Killer\r\n" +
		"\n" +
		"1. Kate Bush – Running Up That Hill\n" +
		"Seal, Kiss From a Rose\n" + // no divider
		"AC/DC - Back In Black - Remastered\n" +
		"Prince\tKiss\n" +
		"Alt-J - Breezeblocks\n" +
		" - No Artist\n"
	got, skipped := ParseLines(text)
	want := []Song{
		{Artist: "ATB", Title: "Killer"},
		{Artist: "Kate Bush", Title: "Running Up That Hill"},
		{Artist: "AC/DC", Title: "Back In Black - Remastered"},
		{Artist: "Prince", Title: "Kiss"},
		{Artist: "Alt-J", Title: "Breezeblocks"},
	}
	if !reflect.DeepEqual(got, want) || skipped != 2 {
		t.Errorf("got %+v, skipped %d; want %+v, skipped 2", got, skipped, want)
	}
}
