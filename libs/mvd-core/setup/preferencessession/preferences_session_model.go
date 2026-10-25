package preferencessession

import (
	"strconv"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/settings"
)

// The settings the screen lists, in order.
const (
	itemOutputDir = iota
	itemVideoQuality
	itemAudioQuality
	itemMergeFormat
	itemTemplate
	itemConcurrent
	itemOfficial
	itemCookies
	itemLogFile
	itemLogDir
	itemSaveNotFound
)

// ItemCount is how many settings the screen lists.
const ItemCount = 11

// Labels are what the settings are called, in the order the screen lists them.
var Labels = []string{
	"Output Folder", "Video Quality", "Audio Quality", "Merge Format",
	"Output Template", "Max Concurrent Downloads", "Download Official Music Video",
	"Cookies from Browser", "Create Log File", "Log File Location",
	"Save Songs Without Official Video (CSV)",
}

var mergeOptions = []string{"mp4", "mkv", "webm"}

// Messages the screen says of what the host's questions came to.
const (
	waitingForAnswersNotice = "Answer the questions in the window that opened..."
)

// Session is the preferences screen: a selectable list of settings, each edited in place by a
// toggle, a number, a text, a choice or a folder.
type Session struct {
	// Cfg is the settings as they stand, including changes not yet saved.
	Cfg config.Config
	// Cursor is the setting the person is on.
	Cursor int
	// Mode is how the setting of the cursor is being changed; EditNone when it is not.
	Mode settings.Mode
	// Text is what is typed while a text or a number is edited.
	Text string
	// RadioOpts and RadioIdx are the choices and the chosen one while a choice is edited.
	RadioOpts []string
	RadioIdx  int
	// Status is what happened last, which the screen shows in place of the keys.
	Status string

	// What the host can do for the screen; the screen offers only that.
	CanPickFolder bool
	CanUninstall  bool
	CanMove       bool
}

// New is the session for the settings cfg.
func New(cfg config.Config) Session { return Session{Cfg: cfg} }

// Editing says whether a setting is being changed.
func (s Session) Editing() bool { return s.Mode != settings.EditNone }

// OnTemplate says whether the cursor is on the setting that takes a template, which has a hint.
func (s Session) OnTemplate() bool { return s.Cursor == itemTemplate }

// Display is what setting i reads as.
func (s Session) Display(i int) string {
	c := s.Cfg
	switch i {
	case itemOutputDir:
		return c.OutputDir
	case itemVideoQuality:
		return c.VideoQuality
	case itemAudioQuality:
		return c.AudioQuality
	case itemMergeFormat:
		return c.MergeOutputFormat
	case itemTemplate:
		return c.OutputTemplate
	case itemConcurrent:
		return strconv.Itoa(c.MaxConcurrentDownloads)
	case itemOfficial:
		return string(c.OfficialVideo)
	case itemCookies:
		return cookieLabel(c)
	case itemLogFile:
		return settings.YesNo(c.CreateLogFile)
	case itemLogDir:
		return c.LogDir
	case itemSaveNotFound:
		if c.OfficialVideo != config.OfficialOnly {
			return settings.YesNo(c.SaveNotFound) + " (used when Download Official is only)"
		}

		return settings.YesNo(c.SaveNotFound)
	}

	return ""
}

// radioState is the choices and the chosen one for the setting of the cursor.
func (s Session) radioState() ([]string, int) {
	switch s.Cursor {
	case itemVideoQuality:
		return config.VideoPresets, indexOf(config.VideoPresets, s.Cfg.VideoQuality)
	case itemAudioQuality:
		return config.AudioPresets, indexOf(config.AudioPresets, s.Cfg.AudioQuality)
	case itemMergeFormat:
		return mergeOptions, indexOf(mergeOptions, s.Cfg.MergeOutputFormat)
	case itemOfficial:
		return config.OfficialModes, indexOf(config.OfficialModes, string(s.Cfg.OfficialVideo))
	case itemCookies:
		return CookieOptions, indexOf(CookieOptions, CurrentCookieChoice(s.Cfg))
	}

	return nil, 0
}

// textValue is the text the setting of the cursor holds.
func (s Session) textValue() string {
	switch s.Cursor {
	case itemOutputDir:
		return s.Cfg.OutputDir
	case itemTemplate:
		return s.Cfg.OutputTemplate
	case itemLogDir:
		return s.Cfg.LogDir
	}

	return ""
}

func indexOf(opts []string, v string) int {
	for i, o := range opts {
		if o == v {
			return i
		}
	}

	return 0
}
