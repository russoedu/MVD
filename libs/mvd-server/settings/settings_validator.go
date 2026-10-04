package settings

import (
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"youtube-downloader/libs/mvd-core/config"
)

// MergeFormats are the containers a merged download can be written as.
var MergeFormats = []string{"mp4", "mkv", "webm"}

// maxWorkers caps both kinds of concurrency. Past this a site rate-limits long before
// the machine is busy.
const maxWorkers = 32

// browserSpec is a yt-dlp browser name with an optional profile. It must start with a
// letter, so a value can never be read by yt-dlp as one of its own flags.
var browserSpec = regexp.MustCompile(`^[a-z][a-z0-9]*(:[^\x00-\x1f]+)?$`)

// Validate says what is wrong with s, or returns nil. Nothing is checked against the
// disk: a folder that does not exist yet is created when it is first needed.
func Validate(s Settings) Invalid {
	bad := Invalid{}

	if msg := folderProblem(s.OutputDir); msg != "" {
		bad["outputDir"] = msg
	}
	if s.CreateLogFile {
		if msg := folderProblem(s.LogDir); msg != "" {
			bad["logDir"] = msg
		}
	}
	if !slices.Contains(config.VideoPresets, s.VideoQuality) {
		bad["videoQuality"] = "choose one of: " + strings.Join(config.VideoPresets, ", ")
	}
	if !slices.Contains(config.AudioPresets, s.AudioQuality) {
		bad["audioQuality"] = "choose one of: " + strings.Join(config.AudioPresets, ", ")
	}
	if !slices.Contains(MergeFormats, s.MergeOutputFormat) {
		bad["mergeOutputFormat"] = "choose one of: " + strings.Join(MergeFormats, ", ")
	}
	if strings.TrimSpace(s.OutputTemplate) == "" || strings.ContainsAny(s.OutputTemplate, "\x00\r\n") {
		bad["outputTemplate"] = "enter a file name template on one line"
	}
	if s.MaxConcurrentDownloads < 1 || s.MaxConcurrentDownloads > maxWorkers {
		bad["maxConcurrentDownloads"] = "enter a number from 1 to 32"
	}
	if s.ConcurrentFragments < 0 || s.ConcurrentFragments > maxWorkers {
		bad["concurrentFragments"] = "enter a number from 0 (off) to 32"
	}
	if s.Cookies != "all" && s.Cookies != "off" && !browserSpec.MatchString(s.Cookies) {
		bad["cookies"] = "choose all, off, or a browser name"
	}

	if len(bad) == 0 {
		return nil
	}

	return bad
}

// folderProblem explains why path cannot be a download or log folder, or returns "".
func folderProblem(path string) string {
	switch {
	case strings.TrimSpace(path) == "":
		return "choose a folder"
	case strings.ContainsAny(path, "\x00\r\n"):
		return "the folder name has characters that cannot be in a path"
	case !filepath.IsAbs(path):
		return "enter the full path of the folder"
	}

	return ""
}
