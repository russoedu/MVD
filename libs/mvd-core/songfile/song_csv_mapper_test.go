package songfile

import (
	"reflect"
	"testing"
)

func TestParseCSVExportify(t *testing.T) {
	text := "\ufeff\"Track URI\",\"Track Name\",\"Artist Name(s)\",\"Duration (ms)\"\n" +
		"\"spotify:track:1\",\"Uptown Girl\",\"Billy Joel\",\"197000\"\n" +
		"\"spotify:track:2\",\"Under Pressure, Live\",\"Queen;David Bowie\",\"248000\"\n" +
		",,,\n" +
		"\"spotify:track:3\",\"No Artist\",\"\",\"1000\"\n"
	got, skipped, ok := ParseCSV(text)
	want := []Song{
		{Title: "Uptown Girl", Artist: "Billy Joel", DurationMs: 197000},
		{Title: "Under Pressure, Live", Artist: "Queen, David Bowie", DurationMs: 248000},
	}
	if !ok || !reflect.DeepEqual(got, want) || skipped != 1 {
		t.Errorf("got %+v, skipped %d, ok %v; want %+v, skipped 1", got, skipped, ok, want)
	}
}

func TestParseCSVDelimitersAndDurations(t *testing.T) {
	semicolons := "Title;Artist;Length\nKiller;ATB;4:08\nKiss;Prince;225\n"
	got, _, ok := ParseCSV(semicolons)
	want := []Song{{Title: "Killer", Artist: "ATB", DurationMs: 248000}, {Title: "Kiss", Artist: "Prince", DurationMs: 225000}}
	if !ok || !reflect.DeepEqual(got, want) {
		t.Errorf("semicolons: got %+v, %v; want %+v", got, ok, want)
	}

	tabs := "artist\ttrack\nSeal\tKiss From a Rose\n"
	got, _, ok = ParseCSV(tabs)
	if !ok || len(got) != 1 || got[0].Artist != "Seal" || got[0].Title != "Kiss From a Rose" {
		t.Errorf("tabs: got %+v, %v", got, ok)
	}
}

func TestParseCSVNeedsTitleAndArtistColumns(t *testing.T) {
	for _, text := range []string{"", "Artist - Title\nATB - Killer\n", "Title,Album\nKiller,Two\n"} {
		if _, _, ok := ParseCSV(text); ok {
			t.Errorf("%q has no title and artist columns", text)
		}
	}
}
