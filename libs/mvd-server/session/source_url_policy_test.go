package session

import (
	"reflect"
	"testing"
)

func TestNormalizeURLsAcceptsHttpAndHttpsAndTrims(t *testing.T) {
	accepted, rejected := NormalizeURLs([]string{
		"  https://www.youtube.com/playlist?list=PL1  ",
		"http://example.com/v",
	})

	want := []string{"https://www.youtube.com/playlist?list=PL1", "http://example.com/v"}
	if !reflect.DeepEqual(accepted, want) || len(rejected) != 0 {
		t.Errorf("accepted %v rejected %v, want %v and none", accepted, rejected, want)
	}
}

func TestNormalizeURLsIgnoresBlankLinesAndCommentsSilently(t *testing.T) {
	accepted, rejected := NormalizeURLs([]string{"", "   ", "# my favourites", "https://youtu.be/abc"})

	if !reflect.DeepEqual(accepted, []string{"https://youtu.be/abc"}) || len(rejected) != 0 {
		t.Errorf("accepted %v rejected %v", accepted, rejected)
	}
}

func TestNormalizeURLsReportsWhatItDoesNotUnderstand(t *testing.T) {
	_, rejected := NormalizeURLs([]string{
		"hello",
		"ftp://example.com/file",
		"//example.com/no-scheme",
		"https://",
		"javascript:alert(1)",
		"file:///etc/passwd",
	})

	want := []string{"hello", "ftp://example.com/file", "//example.com/no-scheme", "https://", "javascript:alert(1)", "file:///etc/passwd"}
	if !reflect.DeepEqual(rejected, want) {
		t.Errorf("rejected %v, want %v", rejected, want)
	}
}

func TestNormalizeURLsKeepsARepeatedURLOnceInOrder(t *testing.T) {
	accepted, _ := NormalizeURLs([]string{"https://a.example/1", "https://b.example/2", "https://a.example/1"})

	if !reflect.DeepEqual(accepted, []string{"https://a.example/1", "https://b.example/2"}) {
		t.Errorf("accepted %v", accepted)
	}
}

func TestNormalizeURLsOfNothingIsNothing(t *testing.T) {
	accepted, rejected := NormalizeURLs(nil)

	if len(accepted) != 0 || len(rejected) != 0 {
		t.Errorf("accepted %v rejected %v", accepted, rejected)
	}
}
