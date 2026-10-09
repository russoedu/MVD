package songfile

import (
	"encoding/csv"
	"strconv"
	"strings"
)

var (
	titleColumns    = []string{"title", "track name", "track title", "track", "song", "song name", "name"}
	artistColumns   = []string{"artist", "artists", "artist name", "artist names", "artist name(s)", "artist(s)"}
	durationMsNames = []string{"duration (ms)", "duration_ms", "duration ms", "length (ms)"}
	durationNames   = []string{"duration", "length", "time"}
)

// ParseCSV reads a CSV whose first row names its columns: a title column, an
// artist column and, if there is one, a duration. The delimiter (comma,
// semicolon or tab) is found from the header. It returns false when the
// header has no title or no artist column, so the caller can read the text as
// lines instead. Rows without a title or an artist are counted as skipped.
func ParseCSV(text string) (songs []Song, skipped int, ok bool) {
	text = strings.TrimPrefix(text, "\ufeff")
	reader := csv.NewReader(strings.NewReader(text))
	reader.Comma = delimiter(text)
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	rows, err := reader.ReadAll()
	if err != nil || len(rows) == 0 {
		return nil, 0, false
	}

	header := rows[0]
	title := column(header, titleColumns)
	artist := column(header, artistColumns)
	if title < 0 || artist < 0 {
		return nil, 0, false
	}
	durationMs := column(header, durationMsNames)
	duration := column(header, durationNames)

	for _, row := range rows[1:] {
		song := Song{Title: field(row, title), Artist: artistList(field(row, artist))}
		if song.Title == "" || song.Artist == "" {
			if !blank(row) {
				skipped++
			}
			continue
		}
		switch {
		case durationMs >= 0:
			song.DurationMs = parseMillis(field(row, durationMs))
		case duration >= 0:
			song.DurationMs = parseLength(field(row, duration))
		}
		songs = append(songs, song)
	}
	return songs, skipped, true
}

// delimiter picks the separator the header line uses most.
func delimiter(text string) rune {
	first, _, _ := strings.Cut(text, "\n")
	best, count := ',', strings.Count(first, ",")
	for _, candidate := range []rune{';', '\t'} {
		if n := strings.Count(first, string(candidate)); n > count {
			best, count = candidate, n
		}
	}
	return best
}

// column returns the index of the first header that is one of the names.
func column(header, names []string) int {
	for _, name := range names {
		for i, h := range header {
			if strings.EqualFold(strings.TrimSpace(h), name) {
				return i
			}
		}
	}
	return -1
}

func field(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func blank(row []string) bool {
	for _, f := range row {
		if strings.TrimSpace(f) != "" {
			return false
		}
	}
	return true
}

// artistList turns the semicolons some exports put between artists into commas.
func artistList(value string) string {
	parts := strings.Split(value, ";")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return strings.Join(parts, ", ")
}

func parseMillis(value string) int {
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// parseLength reads "3:45", "1:02:03" or plain seconds (milliseconds when the
// number is too big to be seconds of a song).
func parseLength(value string) int {
	if strings.Contains(value, ":") {
		seconds := 0
		for _, part := range strings.Split(value, ":") {
			n, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil || n < 0 {
				return 0
			}
			seconds = seconds*60 + n
		}
		return seconds * 1000
	}
	n := parseMillis(value)
	if n >= 10000 {
		return n
	}
	return n * 1000
}
