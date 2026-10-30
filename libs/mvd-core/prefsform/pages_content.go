package prefsform

import (
	"runtime"
	"strconv"
	"strings"

	"youtube-downloader/libs/mvd-core/config"
)

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

var cookieOptions = []string{"off", "mvd", "firefox", "chrome", "edge", "brave", "chromium", "opera", "vivaldi", "safari", "all"}

// signInLabel names the sign-in button, with what the last sign-in came to.
func signInLabel(state string) string {
	switch state {
	case "premium":
		return "Sign in to YouTube (signed in, Premium)"
	case "free":
		return "Sign in to YouTube (signed in, not Premium)"
	case "none":
		return "Sign in to YouTube (no login found)"
	}

	return "Sign in to YouTube"
}

// cookieOptionsFor leaves out the sign-in made inside MVD where the host cannot offer it.
func cookieOptionsFor(caps Capabilities) []string {
	if caps.CanSignIn {
		return cookieOptions
	}
	var options []string
	for _, o := range cookieOptions {
		if o != "mvd" {
			options = append(options, o)
		}
	}

	return options
}

func showCookieOption(option string) string {
	if option == "mvd" {
		return "MVD sign-in"
	}

	return capitalise(option)
}

// colourField is a theme colour; slot points at it inside the config.
func colourField(label string, slot func(*config.ThemeColors) *string) Field {
	return Field{
		Label: label, Kind: KindColour, Slot: slot,
		Get: func(c config.Config) string { return *slot(&c.Colors) },
		// Changing a colour makes the theme the person's own, with the one before as its base.
		Set: func(c *config.Config, v string) { *slot(&c.Colors) = v; c.ThemeName = config.CustomTheme },
	}
}

// windowStyleField is the choice of how the window looks: MVD's terminal look, or one of the
// classic systems.
func windowStyleField() Field {
	return Field{
		Label: "Look", Kind: KindChoice, Options: config.WindowStyleIDs(), Show: config.WindowStyleName,
		Hint: "The terminal look, or the look of a classic Windows or Mac. A change applies at once",
		Get: func(c config.Config) string {
			if config.ValidWindowStyle(c.WindowStyle) {
				return c.WindowStyle
			}
			return config.WindowStyleDOS
		},
		// Another style starts with its own first scheme.
		Set: func(c *config.Config, v string) { c.WindowStyle, c.WindowScheme = v, "" },
	}
}

// classicStyle says whether the settings draw the window in a classic system's look.
func classicStyle(c config.Config) bool {
	return config.ValidWindowStyle(c.WindowStyle) && c.WindowStyle != config.WindowStyleDOS
}

// windowSchemeField is the colour scheme of the classic style chosen, as the system offered them.
func windowSchemeField() Field {
	return Field{
		Label: "Colour scheme", Kind: KindChoice,
		Hint:     "The colours of the window, as the system offered them",
		Choices:  func(c config.Config) []string { return config.WindowSchemeIDs(c.WindowStyle) },
		ShowWith: func(c config.Config, v string) string { return config.WindowSchemeName(c.WindowStyle, v) },
		Shown:    classicStyle,
		Get:      func(c config.Config) string { return config.WindowSchemeOf(c.WindowStyle, c.WindowScheme) },
		Set:      func(c *config.Config, v string) { c.WindowScheme = v },
	}
}

// themeField is the choice of theme: each of the themes of the app, and the person's own.
func themeField() Field {
	return Field{
		Label: "Colours", Kind: KindChoice, Options: config.ThemeNames(),
		Hint: "The terminal look's colours: pick a set to start from, then change any colour below to make it yours",
		Get:  func(c config.Config) string { return c.ThemeName },
		Set: func(c *config.Config, v string) {
			if colours, ok := config.PresetColors(v); ok {
				c.Colors = colours
			}
			c.ThemeName = v
		},
		Swatches: func(option string) []string {
			c, ok := config.PresetColors(option)
			if !ok {
				return nil
			}
			return []string{c.Background, c.Accent, c.Focus, c.Highlight, c.Success, c.Error}
		},
	}
}

// Pages are the tabs, in the order they are shown.
func Pages(caps Capabilities) []Page {
	advanced := []Field{
		{Label: "Parts downloaded at once", Hint: "Each video is fetched in pieces; this is how many come at once. More can be faster.", Kind: KindStepper, Min: 0, Max: 16, Step: 1, Off: "off",
			Get: func(c config.Config) string { return intText(c.ConcurrentFragments) },
			Set: func(c *config.Config, v string) { c.ConcurrentFragments = parseInt(v) }},
		{Label: "Give up on a slow start after (seconds)", Kind: KindStepper, Hint: "How long a download may take to start before it is stopped.", Min: 10, Max: 3600, Step: 10,
			Get: func(c config.Config) string { return intText(c.StartTimeoutSeconds) },
			Set: func(c *config.Config, v string) { c.StartTimeoutSeconds = parseInt(v) }},
		{Label: "Use my browser's YouTube login", Hint: "Off reads nothing from your browsers. Pick one to use its login for age-restricted or private videos; All tries each installed browser.", Kind: KindChoice, Options: cookieOptionsFor(caps), Show: showCookieOption,
			Get: cookieChoice, Set: setCookieChoice},
		{Label: "List songs with no official video (CSV)", Hint: "Only matters when Download Official Music Video is set to Only: keeps a CSV list of the songs no official video was found for.", Kind: KindChoice, Options: yesNo,
			Get: func(c config.Config) string { return boolText(c.SaveNotFound) },
			Set: func(c *config.Config, v string) { c.SaveNotFound = v == "yes" }},
		{Label: "Create Log File", Hint: "Write a log of what the app did, which helps to find out what went wrong.", Kind: KindChoice, Options: yesNo,
			Get: func(c config.Config) string { return boolText(c.CreateLogFile) },
			Set: func(c *config.Config, v string) { c.CreateLogFile = v == "yes" }},
		{Label: "Log File Location", Hint: "The folder the log is written to.", Kind: KindFolder,
			Get: func(c config.Config) string { return c.LogDir },
			Set: func(c *config.Config, v string) { c.LogDir = v }},
		{Label: "Check for Updates", Hint: "Once a day, ask whether a newer MVD is out. Nothing is downloaded until you say Update now.", Kind: KindChoice, Options: yesNo,
			Get: func(c config.Config) string { return boolText(c.CheckForUpdates) },
			Set: func(c *config.Config, v string) { c.CheckForUpdates = v == "yes" }},
	}
	if caps.CanSignIn {
		advanced = append(advanced, Field{Label: signInLabel(caps.SignInState), Hint: "Opens a window to sign in once. With a Premium account the music downloads at a higher quality. Then pick MVD sign-in in the setting above.", Kind: KindAction, Action: ActionSignIn})
	}
	if caps.CanMove {
		advanced = append(advanced, Field{Label: "Move MVD to its own folder", Kind: KindAction, Action: ActionMove})
	}
	if caps.CanUninstall {
		advanced = append(advanced, Field{Label: "Uninstall MVD", Kind: KindAction, Action: ActionUninstall})
	}
	theme := []Field{
		themeField(),
		func() Field {
			f := colourField("Background", func(t *config.ThemeColors) *string { return &t.Background })
			f.Optional = true
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
	}
	if caps.CanStyleWindow {
		// The terminal look's themes and colours are for the terminal look only.
		terminalOnly := func(c config.Config) bool { return !classicStyle(c) }
		for i := range theme {
			theme[i].Shown = terminalOnly
		}
		theme = append([]Field{windowStyleField(), windowSchemeField()}, theme...)
	}
	return []Page{
		{Name: "Output", Fields: []Field{
			{Label: "Videos Output Folder", Hint: "The folder your videos are saved in. Browse… picks another one.", Kind: KindFolder,
				Get: func(c config.Config) string { return c.OutputDir },
				Set: func(c *config.Config, v string) { c.OutputDir = v }},
			{Label: "Output Template", Hint: "How each saved file is named and filed. Add a block such as ARTIST or TITLE, and type any text between them. The way it is now works; change it only if you want to.", Kind: KindTemplate,
				Get: func(c config.Config) string { return c.OutputTemplate },
				Set: func(c *config.Config, v string) { c.OutputTemplate = v }},
		}},
		{Name: "Download", Fields: []Field{
			{Label: "Video Quality", Hint: "The best picture quality to download. A lower one makes smaller files.", Kind: KindChoice, Options: config.VideoPresets,
				Get: func(c config.Config) string { return c.VideoQuality },
				Set: func(c *config.Config, v string) { c.VideoQuality = v }},
			{Label: "Audio Quality", Hint: "The sound quality to keep.", Kind: KindChoice, Options: config.AudioPresets,
				Get: func(c config.Config) string { return c.AudioQuality },
				Set: func(c *config.Config, v string) { c.AudioQuality = v }},
			{Label: "Merge Format", Hint: "The kind of video file made. MP4 plays almost everywhere; MKV and WebM can hold higher quality.", Kind: KindChoice, Options: []string{"mp4", "mkv", "webm"}, Show: upper,
				Get: func(c config.Config) string { return c.MergeOutputFormat },
				Set: func(c *config.Config, v string) { c.MergeOutputFormat = v }},
			{Label: "Download Official Music Video", Hint: "Yes: use the official music video when there is one, and the video as it is otherwise. Only: download just the songs that have one. No: never look for it.", Kind: KindChoice, Options: config.OfficialModes,
				Get: func(c config.Config) string { return string(c.OfficialVideo) },
				Set: func(c *config.Config, v string) { c.OfficialVideo = config.OfficialMode(v) }},
			{Label: "Best Audio On The Video", Hint: "The video is recreated with the best possible music MVD can get: after an official music video is downloaded, the best quality of the song is fetched and put in place of its sound, lined up with the picture. It takes extra time for every song, and the video you downloaded is kept if anything goes wrong. Windows only.", Kind: KindChoice, Options: yesNo,
				Shown: func(config.Config) bool { return runtime.GOOS == "windows" },
				Get:   func(c config.Config) string { return boolText(c.BetterAudio) },
				Set:   func(c *config.Config, v string) { c.BetterAudio = v == "yes" }},
			{Label: "Max Concurrent Downloads", Hint: "How many videos are downloaded at the same time. More is faster but uses more of your connection.", Kind: KindStepper, Min: 1, Max: 32, Step: 1,
				Get: func(c config.Config) string { return intText(c.MaxConcurrentDownloads) },
				Set: func(c *config.Config, v string) { c.MaxConcurrentDownloads = parseInt(v) }},
			{Label: "Retry Failed Downloads", Hint: "Try a download again by itself when it fails, because of a dropped connection, for example.", Kind: KindChoice, Options: yesNo,
				Get: func(c config.Config) string { return boolText(c.AutoRetry) },
				Set: func(c *config.Config, v string) { c.AutoRetry = v == "yes" }},
			{Label: "Retry Attempts", Hint: "How many times a failed download is tried again before it is given up.", Kind: KindStepper, Min: 1, Max: 999, Step: 1,
				Get: func(c config.Config) string { return intText(c.RetryCycles) },
				Set: func(c *config.Config, v string) { c.RetryCycles = parseInt(v) }},
			{Label: "Wait before retrying (seconds)", Kind: KindStepper, Hint: "The wait before the first retry, times the number of the attempt.", Min: 1, Max: 3600, Step: 5,
				Get: func(c config.Config) string { return intText(c.RetryWaitSeconds) },
				Set: func(c *config.Config, v string) { c.RetryWaitSeconds = parseInt(v) }},
		}},
		{Name: "Advanced", Fields: advanced},
		{Name: "Theme", Fields: theme},
	}
}
