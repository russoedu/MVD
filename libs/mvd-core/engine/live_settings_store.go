package engine

import "time"

// LiveSettings are the options of a run that can change while it goes on. Each is read
// again every time it is used, so a change reaches the next download (or the next
// retry, or the next lookup) and leaves the ones under way as they are.
type LiveSettings struct {
	// Workers is the most downloads in flight; the throttle may allow fewer while the
	// service limits requests. Raising it takes effect at once, lowering it as the
	// downloads under way finish.
	Workers int

	Quality             string
	MergeOutputFormat   string
	OutputDir           string
	OutputTemplate      string
	ConcurrentFragments int
	ExtraArgs           []string

	AutoRetry    bool
	RetryCycles  int
	RetryWait    time.Duration
	StartTimeout time.Duration

	// LookForOfficial turns the lookup of official videos on or off, for the songs not yet
	// looked up; OfficialOnly ends the songs that have none as not found.
	LookForOfficial bool
	OfficialOnly    bool
}

// liveOf is the live settings a run starts with: what its options say.
func liveOf(o Options) LiveSettings {
	return LiveSettings{
		Workers:             atLeast(o.Workers, 1),
		Quality:             o.Quality,
		MergeOutputFormat:   o.MergeOutputFormat,
		OutputDir:           o.OutputDir,
		OutputTemplate:      o.OutputTemplate,
		ConcurrentFragments: o.ConcurrentFragments,
		ExtraArgs:           append([]string(nil), o.ExtraArgs...),
		AutoRetry:           o.AutoRetry,
		RetryCycles:         o.RetryCycles,
		RetryWait:           o.RetryWait,
		StartTimeout:        o.StartTimeout,
		LookForOfficial:     o.Resolver != nil && !o.OfficialLookupOff,
		OfficialOnly:        o.OfficialOnly,
	}
}

// OfficialOnlySetter is what a resolver offers when it can be told, while it works,
// whether to leave out the choice of the best quality upload.
type OfficialOnlySetter interface {
	SetOfficialOnly(bool)
}

// Settings returns the live settings as they stand.
func (e *Engine) Settings() LiveSettings {
	return e.settings()
}

func (e *Engine) settings() LiveSettings {
	e.liveMu.RLock()
	defer e.liveMu.RUnlock()
	s := e.live
	s.ExtraArgs = append([]string(nil), s.ExtraArgs...)
	return s
}

// Reconfigure replaces the live settings. A change in the number of downloads reaches the
// throttle at once; the rest is read the next time it is needed.
func (e *Engine) Reconfigure(s LiveSettings) {
	s.Workers = min(atLeast(s.Workers, 1), maxDownloaders)
	e.liveMu.Lock()
	e.live = s
	e.liveMu.Unlock()
	e.throttle.setMax(s.Workers)
	if setter, ok := e.opts.Resolver.(OfficialOnlySetter); ok {
		setter.SetOfficialOnly(s.OfficialOnly)
	}
}

// resolver is the official-video resolver while the lookup is on, nil while it is off.
func (e *Engine) resolver() OfficialResolver {
	if e.opts.Resolver == nil {
		return nil
	}
	e.liveMu.RLock()
	on := e.live.LookForOfficial
	e.liveMu.RUnlock()
	if !on {
		return nil
	}
	return e.opts.Resolver
}
