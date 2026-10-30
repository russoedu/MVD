package engine

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// runLogger appends every line to a log file so nothing is lost when the
// TUI redraws the screen.
type runLogger struct {
	path string
	mu   sync.Mutex
	f    *os.File
}

func newRunLogger(path string) (*runLogger, error) {
	l := &runLogger{path: path}
	if path == "" {
		return l, nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot open log file %s: %w", path, err)
	}
	l.f = f
	_, _ = fmt.Fprintf(f, "\n===== MVD run started %s =====\n", time.Now().Format(time.RFC3339))
	return l, nil
}

// Write appends line to the log with the time and a tag for the playlist.
// Playlist is the zero-based playlist index. Entry is the entry id, and a negative entry tags the playlist alone.
// It does nothing when the logger has no file.
func (l *runLogger) Write(playlist, entry int, line string) {
	if l.f == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	tag := fmt.Sprintf("P%d", playlist+1)
	if entry >= 0 {
		tag = fmt.Sprintf("P%d/E%d", playlist+1, entry)
	}
	_, _ = fmt.Fprintf(l.f, "%s [%s] %s\n", time.Now().Format("15:04:05"), tag, line)
}

// Close closes the log file, if there is one.
func (l *runLogger) Close() {
	if l.f != nil {
		_ = l.f.Close()
	}
}
