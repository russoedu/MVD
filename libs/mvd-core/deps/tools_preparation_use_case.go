package deps

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Preparation is the getting of the tools, done while the window is already open: what it is
// doing, whether it is done, and what went wrong if the program cannot work without a tool.
type Preparation struct {
	done chan struct{}

	mu     sync.Mutex
	status string
	failed map[string]error
}

// PrepareIn does what EnsureIn does, but does not make the caller wait. When nothing is
// missing it is done at once, without asking the network anything; otherwise the missing tools
// are fetched in the background and Status says which. PATH is set before it returns, so
// the app's own copies are found as soon as they are there.
func PrepareIn(binDir string, report Reporter) *Preparation {
	return prepare(newInstaller(binDir, nil), report)
}

func prepare(i *installer, report Reporter) *Preparation {
	p := &Preparation{done: make(chan struct{}), failed: map[string]error{}}
	i.report = func(e Event) {
		p.note(e)
		if report != nil {
			report(e)
		}
	}
	i.putBinOnPath()
	if len(i.missing()) == 0 {
		close(p.done)

		return p
	}
	p.status = "Getting the download tools (one time)..."
	go func() {
		defer close(p.done)
		i.ensure()
		p.finish()
	}()

	return p
}

// note keeps what a step says, for Status and for the tools that could not be installed.
func (p *Preparation) note(e Event) {
	p.mu.Lock()
	defer p.mu.Unlock()
	switch e.Kind {
	case EventMissing:
		p.status = "Getting the download tools (one time): " + strings.Join(e.Names, ", ") + "..."
	case EventDownloading:
		p.status = "Getting " + e.Name + " (one time)..."
	case EventFailed:
		p.failed[e.Name] = e.Err
	}
}

// finish clears what the preparation was doing.
func (p *Preparation) finish() {
	p.mu.Lock()
	p.status = ""
	p.mu.Unlock()
}

// Ready says whether the tools needed to download are there: it is false while they are being
// fetched, and stays false when one that cannot be done without could not be.
func (p *Preparation) Ready() bool {
	select {
	case <-p.done:
		return p.Problem() == ""
	default:
		return false
	}
}

// Status is what is being done now, such as "Getting yt-dlp (one time)...", or "" when nothing is.
func (p *Preparation) Status() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.status
}

// Problem says which tool could not be installed and why, or "" when all that are needed are.
// The JavaScript runtime is not needed to start, so it is not counted.
func (p *Preparation) Problem() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, name := range managedTools {
		if err, failed := p.failed[name]; failed {
			return fmt.Sprintf("Could not get %s: %v. Check the connection and start MVD again.", name, err)
		}
	}

	return ""
}

// Wait returns when the tools are done being prepared, or when ctx ends.
func (p *Preparation) Wait(ctx context.Context) error {
	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ToolPath is where the app's own copy of a tool is (or will be) put in binDir, whether it is
// there yet or not.
func ToolPath(binDir, name string) string {
	dep, ok := platformDeps(runtime.GOOS, runtime.GOARCH)[name]
	if !ok {
		return name
	}
	abs, err := filepath.Abs(binDir)
	if err != nil {
		abs = binDir
	}

	return filepath.Join(abs, dep.FileName)
}

// missing are the tools that have to be fetched before the program can work, found without
// asking the network anything.
func (i *installer) missing() []string {
	dir := i.dir()
	known := loadManifest(dir)
	var missing []string
	for _, name := range managedTools {
		if i.needsInstall(i.available[name], dir, known) {
			missing = append(missing, name)
		}
	}
	_, denoErr := i.lookPath("deno")
	_, nodeErr := i.lookPath("node")
	if denoErr != nil && nodeErr != nil {
		missing = append(missing, "deno")
	}

	return missing
}
