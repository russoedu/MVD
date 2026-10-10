package audiofix

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// With a real ffmpeg: a tone on the right channel only is one-sided, a tone on both is
// not, and the file is never touched.
func TestATrackThatPlaysOnOneSideOnlyIsFound(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is not installed")
	}
	dir := t.TempDir()
	sample := func(name, pan string) string {
		path := filepath.Join(dir, name)
		cmd := exec.Command(ffmpeg, "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "sine=frequency=440:duration=2",
			"-af", pan, "-c:a", "aac", path)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("cannot make the sample: %v\n%s", err, out)
		}
		return path
	}
	oneSided := sample("right.m4a", "pan=stereo|c0=0*c0|c1=c0")
	both := sample("both.m4a", "pan=stereo|c0=c0|c1=c0")

	before, _ := os.ReadFile(oneSided)
	if found, err := OneSided(context.Background(), ffmpeg, oneSided); err != nil || !found {
		t.Errorf("one-sided track: found %v, err %v", found, err)
	}
	if found, err := OneSided(context.Background(), ffmpeg, both); err != nil || found {
		t.Errorf("track that plays on both sides: found %v, err %v", found, err)
	}
	if after, _ := os.ReadFile(oneSided); string(after) != string(before) {
		t.Error("the file must not be changed")
	}
}
