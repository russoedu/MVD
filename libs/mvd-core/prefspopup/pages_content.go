package prefspopup

import (
	"strconv"
	"strings"

	"youtube-downloader/libs/mvd-core/config"
)

// page is a tab: its name and the settings under it.
type page struct {
	name   string
	fields []field
}

var yesNo = []string{"yes", "no"}

func boolText(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func intText(n int) string { return strconv.Itoa(n) }

func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func upper(s string) string { return strings.ToUpper(s) }

func cookieChoice(c config.Config) string {
	switch {
	case c.CookiesFromBrowser != "":
		return c.CookiesFromBrowser
	case c.AutoCookies:
		return "all"
	}
	return "off"
}

func setCookieChoice(c *config.Config, v string) {
	switch v {
	case "all":
		c.AutoCookies, c.CookiesFromBrowser = true, ""
	case "off":
		c.AutoCookies, c.CookiesFromBrowser = false, ""
	default:
		c.AutoCookies, c.CookiesFromBrowser = false, v
	}
}

var cookieOptions = []string{"off", "firefox", "chrome", "edge", "brave", "chromium", "opera", "vivaldi", "safari", "all"}

// colourField is a theme colour; slot points at it inside the config.
func colourField(label string, slot func(*config.ThemeColors) *string) field {
	return field{
		label: label, kind: kindColour, slot: slot,
		get: func(c config.Config) string { return *slot(&c.Colors) },
		// Changing a colour makes the theme the person's own, with the one before as its base.
		set: func(c *config.Config, v string) { *slot(&c.Colors) = v; c.ThemeName = config.CustomTheme },
	}
}

// themeField is the choice of theme: each of the themes of the app, and the person's own.
func themeField() field {
	return field{
		label: "Theme", kind: kindChoice, options: config.ThemeNames(),
		hint: "Pick one to start from, then change any colour below to make it yours",
		get:  func(c config.Config) string { return c.ThemeName },
		set: func(c *config.Config, v string) {
			if colours, ok := config.PresetColors(v); ok {
				c.Colors = colours
			}
			c.ThemeName = v
		},
		swatches: func(option string) []string {
			c, ok := config.PresetColors(option)
			if !ok {
				return nil
			}
			return []string{c.Background, c.Accent, c.Focus, c.Highlight, c.Success, c.Error}
		},
	}
}

// pages are the tabs, in the order they are shown.
func pages(host Host) []page {
	advanced := []field{
		{label: "Parts downloaded at once", hint: "Each video is fetched in pieces; this is how many come at once. More can be faster.", kind: kindStepper, min: 0, max: 16, step: 1, off: "off",
			get: func(c config.Config) string { return intText(c.ConcurrentFragments) },
			set: func(c *config.Config, v string) { c.ConcurrentFragments = parseInt(v) }},
		{label: "Give up on a slow start after (seconds)", kind: kindStepper, hint: "How long a download may take to start before it is stopped.", min: 10, max: 3600, step: 10,
			get: func(c config.Config) string { return intText(c.StartTimeoutSeconds) },
			set: func(c *config.Config, v string) { c.StartTimeoutSeconds = parseInt(v) }},
		{label: "Use my browser's YouTube login", hint: "Off reads nothing from your browsers. Pick one to use its login for age-restricted or private videos; All tries each installed browser.", kind: kindChoice, options: cookieOptions,
			get: cookieChoice, set: setCookieChoice},
		{label: "List songs with no official video (CSV)", hint: "Only matters when Download Official Music Video is set to Only: keeps a CSV list of the songs no official video was found for.", kind: kindChoice, options: yesNo,
			get: func(c config.Config) string { return boolText(c.SaveNotFound) },
			set: func(c *config.Config, v string) { c.SaveNotFound = v == "yes" }},
		{label: "Create Log File", hint: "Write a log of what the app did, which helps to find out what went wrong.", kind: kindChoice, options: yesNo,
			get: func(c config.Config) string { return boolText(c.CreateLogFile) },
			set: func(c *config.Config, v string) { c.CreateLogFile = v == "yes" }},
		{label: "Log File Location", hint: "The folder the log is written to.", kind: kindFolder,
			get: func(c config.Config) string { return c.LogDir },
			set: func(c *config.Config, v string) { c.LogDir = v }},
	}
	if host.Move != nil {
		advanced = append(advanced, field{label: "Move MVD to its own folder", kind: kindAction, action: actionMove})
	}
	if host.Uninstall != nil {
		advanced = append(advanced, field{label: "Uninstall MVD", kind: kindAction, action: actionUninstall})
	}
	return []page{
		{name: "Output", fields: []field{
			{label: "Videos Output Folder", hint: "Where the downloaded videos are saved.", kind: kindFolder,
				get: func(c config.Config) string { return c.OutputDir },
				set: func(c *config.Config, v string) { c.OutputDir = v }},
			{label: "Output Template", hint: "How each file is named, and which folders it goes in. Add blocks, and type any text between them.", kind: kindTemplate,
				get: func(c config.Config) string { return c.OutputTemplate },
				set: func(c *config.Config, v string) { c.OutputTemplate = v }},
		}},
		{name: "Download", fields: []field{
			{label: "Video Quality", hint: "The best picture quality to download. A lower one makes smaller files.", kind: kindChoice, options: config.VideoPresets,
				get: func(c config.Config) string { return c.VideoQuality },
				set: func(c *config.Config, v string) { c.VideoQuality = v }},
			{label: "Audio Quality", hint: "The sound quality to keep.", kind: kindChoice, options: config.AudioPresets,
				get: func(c config.Config) string { return c.AudioQuality },
				set: func(c *config.Config, v string) { c.AudioQuality = v }},
			{label: "Merge Format", hint: "The kind of video file made. MP4 plays almost everywhere; MKV and WebM can hold higher quality.", kind: kindChoice, options: []string{"mp4", "mkv", "webm"}, show: upper,
				get: func(c config.Config) string { return c.MergeOutputFormat },
				set: func(c *config.Config, v string) { c.MergeOutputFormat = v }},
			{label: "Download Official Music Video", hint: "Yes: use the official music video when there is one, and the video as it is otherwise. Only: download just the songs that have one. No: never look for it.", kind: kindChoice, options: config.OfficialModes,
				get: func(c config.Config) string { return string(c.OfficialVideo) },
				set: func(c *config.Config, v string) { c.OfficialVideo = config.OfficialMode(v) }},
			{label: "Max Concurrent Downloads", hint: "How many videos are downloaded at the same time. More is faster but uses more of your connection.", kind: kindStepper, min: 1, max: 32, step: 1,
				get: func(c config.Config) string { return intText(c.MaxConcurrentDownloads) },
				set: func(c *config.Config, v string) { c.MaxConcurrentDownloads = parseInt(v) }},
			{label: "Retry Failed Downloads", hint: "Try a download again by itself when it fails, because of a dropped connection, for example.", kind: kindChoice, options: yesNo,
				get: func(c config.Config) string { return boolText(c.AutoRetry) },
				set: func(c *config.Config, v string) { c.AutoRetry = v == "yes" }},
			{label: "Retry Attempts", hint: "How many times a failed download is tried again before it is given up.", kind: kindStepper, min: 1, max: 999, step: 1,
				get: func(c config.Config) string { return intText(c.RetryCycles) },
				set: func(c *config.Config, v string) { c.RetryCycles = parseInt(v) }},
			{label: "Wait before retrying (seconds)", kind: kindStepper, hint: "The wait before the first retry, times the number of the attempt.", min: 1, max: 3600, step: 5,
				get: func(c config.Config) string { return intText(c.RetryWaitSeconds) },
				set: func(c *config.Config, v string) { c.RetryWaitSeconds = parseInt(v) }},
		}},
		{name: "Advanced", fields: advanced},
		{name: "Theme", fields: []field{
			themeField(),
			func() field {
				f := colourField("Background", func(t *config.ThemeColors) *string { return &t.Background })
				f.optional = true
				return f
			}(),
			colourField("Accent (title)", func(t *config.ThemeColors) *string { return &t.Accent }),
			colourField("Focus (active border)", func(t *config.ThemeColors) *string { return &t.Focus }),
			colourField("Highlight (keys, bars)", func(t *config.ThemeColors) *string { return &t.Highlight }),
			colourField("Success", func(t *config.ThemeColors) *string { return &t.Success }),
			colourField("Error", func(t *config.ThemeColors) *string { return &t.Error }),
			colourField("Dim (borders, hints)", func(t *config.ThemeColors) *string { return &t.Dim }),
			colourField("Text", func(t *config.ThemeColors) *string { return &t.Text }),
			colourField("Selected row (background)", func(t *config.ThemeColors) *string { return &t.Selected }),
		}},
	}
}
