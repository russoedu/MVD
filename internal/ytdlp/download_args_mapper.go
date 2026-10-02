package ytdlp

// DownloadOptions are the settings that shape every yt-dlp download.
type DownloadOptions struct {
	Format            string // -f
	OutputTemplate    string // -o, already joined with the output directory
	MergeOutputFormat string // --merge-output-format, empty to skip
	ExtraArgs         []string
}

// DownloadArgs assembles the yt-dlp command line. The trailing arguments
// (flags and the URL to download) are appended as given.
func DownloadArgs(o DownloadOptions, trailing ...string) []string {
	args := []string{
		"-f", o.Format,
		"-o", o.OutputTemplate,
	}
	if o.MergeOutputFormat != "" {
		args = append(args, "--merge-output-format", o.MergeOutputFormat)
	}
	if len(o.ExtraArgs) > 0 {
		args = append(args, o.ExtraArgs...)
	}
	return append(args, trailing...)
}
