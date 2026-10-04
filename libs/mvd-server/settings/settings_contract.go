// Package settings is how a browser reads and changes the common settings.
//
// It exposes the ones a person changes (folders, quality, concurrency, cookies,
// retries) and leaves the advanced ones (raw yt-dlp format and arguments, the cookie
// file path) as they are in config.conf, so saving here never loses them.
package settings

// Settings is what the browser edits. Cookies is "all" (try every installed browser),
// "off", or one browser spec such as "firefox" or "firefox:default".
type Settings struct {
	OutputDir                  string `json:"outputDir"`
	VideoQuality               string `json:"videoQuality"`
	AudioQuality               string `json:"audioQuality"`
	MergeOutputFormat          string `json:"mergeOutputFormat"`
	OutputTemplate             string `json:"outputTemplate"`
	MaxConcurrentDownloads     int    `json:"maxConcurrentDownloads"`
	ConcurrentFragments        int    `json:"concurrentFragments"`
	DownloadOfficialMusicVideo bool   `json:"downloadOfficialMusicVideo"`
	AutoRetry                  bool   `json:"autoRetry"`
	Cookies                    string `json:"cookies"`
	CreateLogFile              bool   `json:"createLogFile"`
	LogDir                     string `json:"logDir"`
}

// Options are the choices a form can offer, so the page never hard-codes them.
type Options struct {
	VideoQualities []string `json:"videoQualities"`
	AudioQualities []string `json:"audioQualities"`
	MergeFormats   []string `json:"mergeFormats"`
	// Browsers are the ones found on this machine, for pinning cookies to one.
	Browsers []string `json:"browsers"`
}

// Document is what GET returns: the values and the choices for them.
type Document struct {
	Settings Settings `json:"settings"`
	Options  Options  `json:"options"`
}

// Invalid lists what is wrong with a Settings, by field name (the JSON key).
type Invalid map[string]string

func (i Invalid) Error() string { return "the settings are not valid" }
