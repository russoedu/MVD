package ytdlp

import "strconv"

// DownloadOptions are the settings that shape every yt-dlp download.
type DownloadOptions struct {
	Format         string // -f
	OutputTemplate string // -o: joined with the output directory, unless HomeDir is set
	// HomeDir is where the finished files go and TempDir where yt-dlp keeps what it is
	// still downloading or merging (-P home: and -P temp:). With HomeDir set the
	// template is relative to it, and a file only appears there once it is complete.
	HomeDir           string
	TempDir           string
	MergeOutputFormat string // --merge-output-format, empty to skip
	// ConcurrentFragments is yt-dlp's --concurrent-fragments: how many DASH
	// fragments of one video to fetch in parallel. 0 or less skips the flag.
	ConcurrentFragments int
	ExtraArgs           []string
}

// DownloadArgs assembles the yt-dlp command line. The trailing arguments
// (flags and the URL to download) are appended as given.
func DownloadArgs(o DownloadOptions, trailing ...string) []string {
	args := []string{"-f", o.Format}
	if o.HomeDir != "" {
		args = append(args, "-P", "home:"+o.HomeDir)
	}
	if o.TempDir != "" {
		args = append(args, "-P", "temp:"+o.TempDir)
	}
	args = append(args, "-o", o.OutputTemplate)
	if o.MergeOutputFormat != "" {
		args = append(args, "--merge-output-format", o.MergeOutputFormat)
	}
	if o.ConcurrentFragments > 0 {
		args = append(args, "--concurrent-fragments", strconv.Itoa(o.ConcurrentFragments))
	}
	if len(o.ExtraArgs) > 0 {
		args = append(args, o.ExtraArgs...)
	}
	return append(args, trailing...)
}
