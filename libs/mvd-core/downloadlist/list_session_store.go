package downloadlist

import (
	"errors"
	"sync"

	"youtube-downloader/libs/mvd-core/playlistfile"
)

// Session is the download list that is open: the file it is saved to (none until the
// person saves it) and what the file holds. It is safe for use by several goroutines.
type Session struct {
	mu   sync.Mutex
	path string
	file playlistfile.File
	// saved is what was last written, so an unchanged list is not written again.
	saved []byte
}

// New starts a list of the given addresses that is not saved anywhere yet.
func New(sources []string) *Session {
	s := &Session{}
	s.setSources(sources)
	return s
}

// Open reads the list at path.
func Open(path string) (*Session, error) {
	file, err := playlistfile.Load(path)
	if err != nil {
		return nil, err
	}
	s := &Session{path: path, file: file}
	s.saved = encode(file)
	return s, nil
}

// Path is the file the list is saved to, "" while it has none.
func (s *Session) Path() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.path
}

// Sources is the address of each item of the list, in order.
func (s *Session) Sources() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.file.Sources()
}

// File is a copy of what the list holds.
func (s *Session) File() playlistfile.File {
	s.mu.Lock()
	defer s.mu.Unlock()
	return copyOf(s.file)
}

// SetSources replaces the addresses with the given ones, in order. An address that was
// already in the list keeps its songs; the others start empty and the ones left out are
// dropped.
func (s *Session) SetSources(sources []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.setSources(sources)
}

func (s *Session) setSources(sources []string) {
	have := make(map[string]playlistfile.Item, len(s.file.Items))
	for _, item := range s.file.Items {
		have[item.Source] = item
	}
	items := make([]playlistfile.Item, 0, len(sources))
	seen := map[string]bool{}
	for _, src := range sources {
		if src == "" || seen[src] {
			continue
		}
		seen[src] = true
		if item, ok := have[src]; ok {
			items = append(items, item)
			continue
		}
		items = append(items, playlistfile.Item{Source: src})
	}
	s.file.Items = items
}

// Save writes the list to path, which becomes the file the list is saved to from then
// on. An empty path saves to the file the list already has.
func (s *Session) Save(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if path == "" {
		path = s.path
	}
	if path == "" {
		return errors.New("the list has no file yet")
	}
	if err := playlistfile.Save(path, s.file); err != nil {
		return err
	}
	s.path = path
	s.saved = encode(s.file)
	return nil
}

// flush writes the list when it has changed since it was last written and has a file.
func (s *Session) flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.path == "" {
		return nil
	}
	now := encode(s.file)
	if string(now) == string(s.saved) {
		return nil
	}
	if err := playlistfile.Save(s.path, s.file); err != nil {
		return err
	}
	s.saved = now
	return nil
}
