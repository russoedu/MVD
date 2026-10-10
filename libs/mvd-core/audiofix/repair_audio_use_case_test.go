package audiofix

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
)

// With a real ffmpeg: a tone on the right channel only comes out on both.
func TestARightOnlyTrackIsCopiedToBothSides(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is not installed")
	}
	path := filepath.Join(t.TempDir(), "song.m4a")
	make := exec.Command(ffmpeg, "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "sine=frequency=440:duration=2",
		"-af", "pan=stereo|c0=0*c0|c1=c0", "-c:a", "aac", path)
	if out, err := make.CombinedOutput(); err != nil {
		t.Fatalf("cannot make the sample: %v\n%s", err, out)
	}

	fixed, err := Repair(context.Background(), ffmpeg, path)
	if err != nil || !fixed {
		t.Fatalf("fixed %v, err %v", fixed, err)
	}
	levels, err := channelLevels(context.Background(), ffmpeg, path)
	if err != nil {
		t.Fatal(err)
	}
	if _, oneSided := SilentChannel(levels); oneSided || len(levels) != 2 || levels[0] < -45 || levels[1] < -45 {
		t.Errorf("both sides should play now, levels %v", levels)
	}

	again, err := Repair(context.Background(), ffmpeg, path)
	if err != nil || again {
		t.Errorf("a file that plays on both sides is left alone, fixed %v, err %v", again, err)
	}
}
