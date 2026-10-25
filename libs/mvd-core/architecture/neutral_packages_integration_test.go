package architecture

import (
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// neutralMarker is what the package comment of a renderer-neutral package says.
const neutralMarker = "renderer-neutral"

const module = "youtube-downloader/libs/mvd-core/"

// renderers are what a renderer-neutral package must not depend on, directly or through another
// package: the terminal libraries, and the packages of this repository that draw or take input.
var renderers = []string{
	"charm.land/",
	"github.com/charmbracelet/",
	"github.com/mattn/go-runewidth",
	module + "look",
	module + "cells",
	module + "screenview",
	module + "keybar",
	module + "pointer",
	module + "dialog",
	module + "menubar",
	module + "lineedit",
	module + "tui",
	module + "internal/",
}

func TestRendererNeutralPackagesDependOnNoRenderer(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	packages := neutralPackages(t, filepath.Join(root, "libs", "mvd-core"))
	if len(packages) == 0 {
		t.Fatal("no package says it is renderer-neutral; the guard would check nothing")
	}
	for _, dir := range packages {
		rel, err := filepath.Rel(root, dir)
		if err != nil {
			t.Fatal(err)
		}
		pattern := "./" + filepath.ToSlash(rel)
		cmd := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}}", pattern)
		cmd.Dir = root
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("go list %s: %v", pattern, err)
		}
		for _, dep := range strings.Fields(string(out)) {
			for _, banned := range renderers {
				if strings.HasPrefix(dep, banned) {
					t.Errorf("%s is renderer-neutral but depends on %s", pattern, dep)
				}
			}
		}
	}
}

// neutralPackages are the directories under dir whose package comment says they are renderer-neutral.
func neutralPackages(t *testing.T, dir string) []string {
	t.Helper()
	var found []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		file, perr := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments|parser.PackageClauseOnly)
		if perr != nil || file.Doc == nil || !strings.Contains(file.Doc.Text(), neutralMarker) {
			return nil
		}
		parent := filepath.Dir(path)
		for _, f := range found {
			if f == parent {
				return nil
			}
		}
		found = append(found, parent)

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	return found
}
