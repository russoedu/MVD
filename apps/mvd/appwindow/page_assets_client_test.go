package appwindow

import (
	"io/fs"
	"strings"
	"testing"
)

func TestThePageIsEmbeddedAndStartsFromAnIndex(t *testing.T) {
	index, err := fs.ReadFile(pageAssets(), "index.html")
	if err != nil {
		t.Fatalf("the built page is missing (run stage-web): %v", err)
	}
	if !strings.Contains(string(index), "<html") {
		t.Errorf("index.html is not a page: %q", index)
	}
}
