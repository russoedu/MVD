// Command youtube-check looks up the official video of every song of a playlist kept for
// the purpose, against the real YouTube, and compares the answers with the ones recorded
// earlier. The official video lookup depends on pages and endpoints YouTube can change
// without notice; when many songs now give a different answer, YouTube (or the lookup)
// changed and the app may need a patch. A few songs differing is normal: videos get
// removed and blocked.
//
// A weekly workflow runs it (.github/workflows/youtube-check.yml). Exit status: 0 when
// the answers still match, 1 when too many differ, 2 when the check could not run.
//
//	youtube-check -playlist <url> -baseline tools/youtube-check/baseline.json
//	youtube-check -playlist <url> -baseline tools/youtube-check/baseline.json -record
package main
