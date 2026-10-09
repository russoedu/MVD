package terminalui

import (
	"context"
	"fmt"
	"os"
	"sync"

	"youtube-downloader/libs/mvd-core/config"
	"youtube-downloader/libs/mvd-core/engine"
	"youtube-downloader/libs/mvd-core/runner"
	"youtube-downloader/libs/mvd-core/tui"
)

// Runs starts the download runs of the terminal interface and remembers the one
// going now, so the app can hear about its failures and try them again once the
// downloader has been updated.
type Runs struct {
	onFailure func()

	mu      sync.Mutex
	current *engineRun
}

// NewRuns prepares a registry; onFailure is called, from the goroutine that
// carries the run's events, each time something in a run fails. It must return quickly.
func NewRuns(onFailure func()) *Runs { return &Runs{onFailure: onFailure} }

// RetryFailed queues the current run's failures again and says how many it queued;
// none when no run is going.
func (r *Runs) RetryFailed() int {
	r.mu.Lock()
	run := r.current
	r.mu.Unlock()
	if run == nil {
		return 0
	}
	return run.ledger.retryAll(run.Engine)
}

// Busy reports whether the current run still has downloads going. A run whose
// downloads are all done is not busy, even while its screen is open.
func (r *Runs) Busy() bool {
	r.mu.Lock()
	run := r.current
	r.mu.Unlock()
	if run == nil {
		return false
	}
	select {
	case <-run.done:
		return false
	default:
		return true
	}
}

// Start is the app model's starter: it builds the engine for the settings and list
// and runs it, until the user leaves its screen or parent ends.
func (r *Runs) Start(parent context.Context, ytDlpPath string, logf func(string, ...interface{})) tui.RunStarter {
	return func(cfg config.Config, urls []string) (tui.Run, error) {
		if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
			return nil, fmt.Errorf("cannot create the download folder %s: %w", cfg.OutputDir, err)
		}
		ctx, cancel := context.WithCancel(parent)
		eng, err := runner.BuildEngine(ctx, ytDlpPath, cfg, urls, logf)
		if err != nil {
			cancel()
			return nil, err
		}
		ledger := newFailureLedger(func(playlist int) string {
			if sources := eng.Sources(); playlist < len(sources) {
				return sources[playlist].URL
			}
			return ""
		})
		run := &engineRun{
			Engine: eng,
			cancel: cancel,
			ledger: ledger,
			events: watchEvents(eng.Events(), ledger, r.onFailure),
			done:   make(chan struct{}),
		}
		run.release = func() {
			r.mu.Lock()
			defer r.mu.Unlock()
			if r.current == run {
				r.current = nil
			}
		}
		go func() { eng.Run(ctx); close(run.done) }()

		r.mu.Lock()
		r.current = run
		r.mu.Unlock()
		return run, nil
	}
}

// engineRun is a download run the app model can close: Close cancels it and waits
// until the engine has stopped and its events are drained, as mvd-tui does. It is
// the engine itself, so the download screen can also add links to it.
type engineRun struct {
	*engine.Engine
	cancel context.CancelFunc
	ledger *failureLedger
	events <-chan interface{}
	done   chan struct{}
	// release forgets the run once it is closed.
	release func()
}

// Events is the engine's events, as the ledger saw them.
func (r *engineRun) Events() <-chan interface{} { return r.events }

func (r *engineRun) Close() {
	r.cancel()
	for range r.events {
	}
	<-r.done
	r.release()
}
