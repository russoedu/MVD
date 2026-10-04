package deps

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// A program that is running cannot be overwritten or deleted on Windows. This is the
// case the update has to handle, because yt-dlp may be in the middle of a download.
func TestAProgramThatIsRunningCanBeReplacedAndTheRunningCopyIsNotDisturbed(t *testing.T) {
	system := os.Getenv("SystemRoot")
	source, err := os.Open(filepath.Join(system, "System32", "ping.exe"))
	if err != nil {
		t.Skip("ping.exe is not available")
	}
	dir := t.TempDir()
	final := filepath.Join(dir, "tool.exe")
	copyOf, err := os.OpenFile(final, os.O_WRONLY|os.O_CREATE, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(copyOf, source); err != nil {
		t.Fatal(err)
	}
	_ = source.Close()
	if err := copyOf.Close(); err != nil {
		t.Fatal(err)
	}

	running := exec.Command(final, "-n", "30", "127.0.0.1")
	if err := running.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = running.Process.Kill()
		_ = running.Wait()
	}()
	time.Sleep(300 * time.Millisecond)

	// Overwriting or deleting it now fails, which is the problem to be solved.
	if err := os.Remove(final); err == nil {
		t.Fatal("expected a running program to be undeletable; the test cannot show anything on this machine")
	}

	partial := final + ".part"
	if err := os.WriteFile(partial, []byte("new program"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := replaceFile(partial, final); err != nil {
		t.Fatalf("a running program could not be replaced: %v", err)
	}

	if read(t, final) != "new program" {
		t.Error("the new program is not in place")
	}
	if running.ProcessState != nil {
		t.Error("the running copy was stopped")
	}
}
