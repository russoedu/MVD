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

var cookieOptions = []string{"all", "off", "firefox", "chrome", "edge", "brave", "chromium", "opera", "vivaldi", "safari"}

// colourField is a theme colour; slot points at it inside the config.
func colourField(label string, slot func(*config.ThemeColors) *string) field {
	return field{
		label: label, kind: kindColour,
		get: func(c config.Config) string { return *slot(&c.Colors) },
		set: func(c *config.Config, v string) { *slot(&c.Colors) = v },
	}
}

// pages are the tabs, in the order they are shown.
func pages(host Host) []page {
	advanced := []field{
		{label: "Parallel Fragments Per Video", kind: kindStepper, min: 0, max: 16, step: 1, off: "off",
			get: func(c config.Config) string { return intText(c.ConcurrentFragments) },
			set: func(c *config.Config, v string) { c.ConcurrentFragments = parseInt(v) }},
		{label: "Extra yt-dlp args", kind: kindText, placeholder: "These are passed directly to yt-dlp...",
			get: func(c config.Config) string { return strings.Join(c.ExtraArgs, " ") },
			set: func(c *config.Config, v string) { c.ExtraArgs = strings.Fields(v) }},
		{label: "Start Timeout Seconds", kind: kindStepper, hint: "How long a download may take to start before it is stopped", min: 10, max: 3600, step: 10,
			get: func(c config.Config) string { return intText(c.StartTimeoutSeconds) },
			set: func(c *config.Config, v string) { c.StartTimeoutSeconds = parseInt(v) }},
		{label: "Cookies from Browser", kind: kindChoice, options: cookieOptions,
			get: cookieChoice, set: setCookieChoice},
		{label: "Save Songs Without Official Video (CSV)", kind: kindChoice, options: yesNo,
			get: func(c config.Config) string { return boolText(c.SaveNotFound) },
			set: func(c *config.Config, v string) { c.SaveNotFound = v == "yes" }},
		{label: "Create Log File", kind: kindChoice, options: yesNo,
			get: func(c config.Config) string { return boolText(c.CreateLogFile) },
			set: func(c *config.Config, v string) { c.CreateLogFile = v == "yes" }},
		{label: "Log File Location", kind: kindFolder,
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
			{label: "Videos Output Folder", kind: kindFolder,
				get: func(c config.Config) string { return c.OutputDir },
				set: func(c *config.Config, v string) { c.OutputDir = v }},
			{label: "Output Template", kind: kindTemplate,
				get: func(c config.Config) string { return c.OutputTemplate },
				set: func(c *config.Config, v string) { c.OutputTemplate = v }},
		}},
		{name: "Download", fields: []field{
			{label: "Video Quality", kind: kindChoice, options: config.VideoPresets,
				get: func(c config.Config) string { return c.VideoQuality },
				set: func(c *config.Config, v string) { c.VideoQuality = v }},
			{label: "Audio Quality", kind: kindChoice, options: config.AudioPresets,
				get: func(c config.Config) string { return c.AudioQuality },
				set: func(c *config.Config, v string) { c.AudioQuality = v }},
			{label: "Merge Format", kind: kindChoice, options: []string{"mp4", "mkv", "webm"}, show: upper,
				get: func(c config.Config) string { return c.MergeOutputFormat },
				set: func(c *config.Config, v string) { c.MergeOutputFormat = v }},
			{label: "Download Official Music Video", kind: kindChoice, options: config.OfficialModes,
				get: func(c config.Config) string { return string(c.OfficialVideo) },
				set: func(c *config.Config, v string) { c.OfficialVideo = config.OfficialMode(v) }},
			{label: "Max Concurrent Downloads", kind: kindStepper, min: 1, max: 32, step: 1,
				get: func(c config.Config) string { return intText(c.MaxConcurrentDownloads) },
				set: func(c *config.Config, v string) { c.MaxConcurrentDownloads = parseInt(v) }},
			{label: "Retry Failed Downloads", kind: kindChoice, options: yesNo,
				get: func(c config.Config) string { return boolText(c.AutoRetry) },
				set: func(c *config.Config, v string) { c.AutoRetry = v == "yes" }},
			{label: "Retry Attempts", kind: kindStepper, min: 1, max: 999, step: 1,
				get: func(c config.Config) string { return intText(c.RetryCycles) },
				set: func(c *config.Config, v string) { c.RetryCycles = parseInt(v) }},
			{label: "Retry Attempts Wait Seconds", kind: kindStepper, hint: "The wait before the first retry, times the number of the attempt", min: 1, max: 3600, step: 5,
				get: func(c config.Config) string { return intText(c.RetryWaitSeconds) },
				set: func(c *config.Config, v string) { c.RetryWaitSeconds = parseInt(v) }},
		}},
		{name: "Advanced", fields: advanced},
		{name: "Theme", fields: []field{
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
