package settings

import (
	"sync"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/cookies"
)

// Repository keeps the settings in config.conf, the file the terminal app reads too.
type Repository struct {
	mu           sync.Mutex
	path         string
	appDir       string
	downloadsDir string
}

// NewRepository returns a repository for the config file at path. appDir and
// downloadsDir are only the defaults used if the file does not exist yet.
func NewRepository(path, appDir, downloadsDir string) *Repository {
	return &Repository{path: path, appDir: appDir, downloadsDir: downloadsDir}
}

// Load returns the current settings and the choices for them.
func (r *Repository) Load() (Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cfg, _, err := config.LoadOrCreate(r.path, r.appDir, r.downloadsDir)
	if err != nil {
		return Document{}, err
	}

	return Document{Settings: From(cfg), Options: choices()}, nil
}

// Save validates s and writes it, keeping every setting this package does not expose.
// If s is not valid it returns an Invalid and writes nothing.
func (r *Repository) Save(s Settings) error {
	if bad := Validate(s); bad != nil {
		return bad
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	cfg, _, err := config.LoadOrCreate(r.path, r.appDir, r.downloadsDir)
	if err != nil {
		return err
	}

	return config.Save(Apply(cfg, s), r.path)
}

func choices() Options {
	browsers := cookies.InstalledBrowsers()
	if browsers == nil {
		browsers = []string{}
	}

	return Options{
		VideoQualities: config.VideoPresets,
		AudioQualities: config.AudioPresets,
		MergeFormats:   MergeFormats,
		Browsers:       browsers,
	}
}
