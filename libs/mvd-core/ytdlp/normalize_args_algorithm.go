package ytdlp

import "strings"

// NormalizeArgs rewrites flags into the form yt-dlp accepts now, so a config written
// for an older yt-dlp keeps working. --js-runtimes used to take a comma separated list
// ("deno,node"); a current yt-dlp reads that as one runtime named "deno,node", ignores
// it, and finds no JavaScript runtime, after which YouTube answers "Video unavailable"
// or 403. Each runtime now needs a flag of its own.
func NormalizeArgs(args []string) []string {
	out := make([]string, 0, len(args)+2)
	for i := 0; i < len(args); i++ {
		switch arg := args[i]; {
		case arg == "--js-runtimes" && i+1 < len(args):
			out = appendRuntimes(out, args[i+1])
			i++
		case strings.HasPrefix(arg, "--js-runtimes="):
			out = appendRuntimes(out, strings.TrimPrefix(arg, "--js-runtimes="))
		default:
			out = append(out, arg)
		}
	}

	return out
}

// appendRuntimes adds one --js-runtimes flag for each runtime in the list.
func appendRuntimes(out []string, list string) []string {
	for _, runtime := range strings.Split(list, ",") {
		if runtime = strings.TrimSpace(runtime); runtime != "" {
			out = append(out, "--js-runtimes", runtime)
		}
	}

	return out
}
