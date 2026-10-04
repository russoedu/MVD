package tray

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type menuProbe struct {
	opens, quits, exits atomic.Int32
	order               chan string
}

func newMenuProbe() *menuProbe { return &menuProbe{order: make(chan string, 8)} }

func (p *menuProbe) open() { p.opens.Add(1); p.order <- "open" }
func (p *menuProbe) quit() { p.quits.Add(1); p.order <- "quit" }
func (p *menuProbe) exit() { p.exits.Add(1); p.order <- "exit" }

func runMenu(ctx context.Context, p *menuProbe) (open, quit chan struct{}, done chan struct{}) {
	open, quit, done = make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() {
		serveTrayMenu(ctx, open, quit, p.open, p.quit, p.exit)
		close(done)
	}()

	return open, quit, done
}

func waitDone(t *testing.T, done chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("serveTrayMenu did not return")
	}
}

func TestQuitStopsTheAppAndThenRemovesTheIcon(t *testing.T) {
	p := newMenuProbe()
	_, quit, done := runMenu(context.Background(), p)

	quit <- struct{}{}
	waitDone(t, done)

	if p.quits.Load() != 1 || p.exits.Load() != 1 {
		t.Fatalf("quit called %d times, exit called %d times; want 1 and 1", p.quits.Load(), p.exits.Load())
	}
	if first, second := <-p.order, <-p.order; first != "quit" || second != "exit" {
		t.Errorf("order = %s then %s, want quit then exit", first, second)
	}
}

func TestOpenCanBeClickedRepeatedlyAndDoesNotEndTheMenu(t *testing.T) {
	p := newMenuProbe()
	open, quit, done := runMenu(context.Background(), p)

	open <- struct{}{}
	open <- struct{}{}
	open <- struct{}{}
	quit <- struct{}{}
	waitDone(t, done)

	if p.opens.Load() != 3 {
		t.Errorf("open called %d times, want 3", p.opens.Load())
	}
}

func TestWhenTheAppStopsByItselfTheIconIsRemovedWithoutCallingQuit(t *testing.T) {
	p := newMenuProbe()
	ctx, cancel := context.WithCancel(context.Background())
	_, _, done := runMenu(ctx, p)

	cancel()
	waitDone(t, done)

	if p.exits.Load() != 1 || p.quits.Load() != 0 {
		t.Errorf("exit called %d times, quit %d times; want 1 and 0", p.exits.Load(), p.quits.Load())
	}
}
