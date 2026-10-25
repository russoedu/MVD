package preferencessession

import (
	"errors"
	"testing"

	"youtube-downloader/libs/mvd-core/settings"
)

func TestAFolderSettingAsksTheHostsChooserWhenThereIsOne(t *testing.T) {
	s := newSession()
	s.CanPickFolder = true
	eff := s.Activate()
	if eff.Kind != PickFolder || eff.Folder != "/dl" || eff.Item != itemOutputDir || s.Mode != settings.EditPicking {
		t.Fatalf("the chooser should be asked for at the current folder: %+v mode %v", eff, s.Mode)
	}
	if eff := s.FolderPicked(eff.Item, "/music", true, nil); eff.Kind != NoEffect {
		t.Errorf("a folder chosen asked the host for %+v", eff)
	}
	if s.Mode != settings.EditNone || s.Cfg.OutputDir != "/music" {
		t.Errorf("mode %v, output dir %q", s.Mode, s.Cfg.OutputDir)
	}
}

func TestCancellingTheChooserChangesNothing(t *testing.T) {
	s := newSession()
	s.CanPickFolder = true
	s.Activate()
	s.FolderPicked(itemOutputDir, "", false, nil)
	if s.Mode != settings.EditNone || s.Cfg.OutputDir != "/dl" {
		t.Errorf("mode %v, output dir %q", s.Mode, s.Cfg.OutputDir)
	}
}

func TestWithoutAChooserOnTheMachineTheBrowserIsAskedFor(t *testing.T) {
	s := newSession()
	s.CanPickFolder = true
	s.Activate()
	eff := s.FolderPicked(itemOutputDir, "", false, errors.New("no desktop"))
	if eff.Kind != BrowseFolders || eff.Folder != "/dl" || s.Mode != settings.EditFolder {
		t.Errorf("effect %+v mode %v, want the browser", eff, s.Mode)
	}
}

func TestWithoutAHostChooserTheBrowserIsAskedForAtOnce(t *testing.T) {
	s := newSession()
	eff := s.Activate()
	if eff.Kind != BrowseFolders || s.Mode != settings.EditFolder {
		t.Errorf("effect %+v mode %v", eff, s.Mode)
	}
}

func TestTheLogFolderUsesTheChooserToo(t *testing.T) {
	s := newSession()
	s.CanPickFolder = true
	s.Cursor = itemLogDir
	eff := s.Activate()
	s.FolderPicked(eff.Item, "/logs", true, nil)
	if s.Cfg.LogDir != "/logs" || s.Cfg.OutputDir != "/dl" {
		t.Errorf("log dir %q output dir %q", s.Cfg.LogDir, s.Cfg.OutputDir)
	}
}

func TestTheBrowserSetsTheFolderOfTheCursorOrChangesNothing(t *testing.T) {
	s := newSession()
	s.Activate()
	s.FolderChosen("/music")
	if s.Mode != settings.EditNone || s.Cfg.OutputDir != "/music" {
		t.Errorf("mode %v output dir %q", s.Mode, s.Cfg.OutputDir)
	}
	s.Cursor = itemLogDir
	s.Activate()
	before := s.Cfg.LogDir
	s.FolderCancelled()
	if s.Mode != settings.EditNone || s.Cfg.LogDir != before {
		t.Errorf("cancelling changed the log folder to %q", s.Cfg.LogDir)
	}
	s.Activate()
	s.FolderChosen("/logs")
	if s.Cfg.LogDir != "/logs" {
		t.Errorf("log dir %q", s.Cfg.LogDir)
	}
}
