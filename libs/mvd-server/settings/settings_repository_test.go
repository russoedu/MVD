package settings

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"youtube-downloader/libs/mvd-core/config"
)

func newRepository(t *testing.T) (*Repository, string) {
	t.Helper()
	appDir := t.TempDir()
	path := filepath.Join(appDir, "config.conf")

	return NewRepository(path, appDir, t.TempDir()), path
}

func TestLoadCreatesTheFileWithDefaultsAndOffersTheChoices(t *testing.T) {
	repo, path := newRepository(t)

	doc, err := repo.Load()

	if err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Errorf("the config file was not created: %v", statErr)
	}
	if doc.Settings.VideoQuality != "best" || doc.Settings.MaxConcurrentDownloads != 4 {
		t.Errorf("defaults = %+v", doc.Settings)
	}
	if len(doc.Options.VideoQualities) == 0 || len(doc.Options.MergeFormats) != 3 || doc.Options.Browsers == nil {
		t.Errorf("options = %+v", doc.Options)
	}
}

func TestSaveThenLoadReturnsWhatWasSaved(t *testing.T) {
	repo, _ := newRepository(t)
	doc, err := repo.Load()
	if err != nil {
		t.Fatal(err)
	}
	want := doc.Settings
	want.OutputDir = t.TempDir()
	want.VideoQuality = "1080p"
	want.MaxConcurrentDownloads = 2
	want.Cookies = "off"

	if err := repo.Save(want); err != nil {
		t.Fatal(err)
	}

	got, err := repo.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Settings, want) {
		t.Errorf("got %+v\nwant %+v", got.Settings, want)
	}
}

func TestSaveKeepsTheAdvancedSettingsThatTheFileHolds(t *testing.T) {
	repo, path := newRepository(t)
	cfg, _, err := config.LoadOrCreate(path, filepath.Dir(path), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cfg.RawFormat = "bv*+ba"
	cfg.ExtraArgs = []string{"--proxy", "http://p"}
	if err := config.Save(cfg, path); err != nil {
		t.Fatal(err)
	}
	doc, _ := repo.Load()
	changed := doc.Settings
	changed.AudioQuality = "low"

	if err := repo.Save(changed); err != nil {
		t.Fatal(err)
	}

	after, _, err := config.LoadOrCreate(path, filepath.Dir(path), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if after.RawFormat != "bv*+ba" || !reflect.DeepEqual(after.ExtraArgs, []string{"--proxy", "http://p"}) || after.AudioQuality != "low" {
		t.Errorf("after = %+v", after)
	}
}

func TestAnInvalidSaveWritesNothingAndSaysWhy(t *testing.T) {
	repo, path := newRepository(t)
	doc, _ := repo.Load()
	before, _ := os.ReadFile(path)
	bad := doc.Settings
	bad.MaxConcurrentDownloads = 0

	err := repo.Save(bad)

	var invalid Invalid
	if !errors.As(err, &invalid) || invalid["maxConcurrentDownloads"] == "" {
		t.Fatalf("err = %v", err)
	}
	after, _ := os.ReadFile(path)
	if string(after) != string(before) {
		t.Error("the file changed although the settings were invalid")
	}
}
