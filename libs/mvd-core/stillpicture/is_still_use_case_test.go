package stillpicture

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// picture is a frame with a bright block at a position, on a grey ground.
func picture(blockX int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 160, 120))
	for y := 0; y < 120; y++ {
		for x := 0; x < 160; x++ {
			c := color.RGBA{R: 60, G: 60, B: 60, A: 255}
			if x >= blockX && x < blockX+50 && y >= 30 && y < 90 {
				c = color.RGBA{R: 250, G: 250, B: 250, A: 255}
			}
			img.Set(x, y, c)
		}
	}
	return img
}

func encoded(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestFramesOfTheSamePictureAreStill(t *testing.T) {
	frames := []image.Image{picture(20), picture(20), picture(20)}
	if !looksStill(frames) {
		t.Error("three identical frames are a still picture")
	}
}

func TestFramesThatMoveAreNotStill(t *testing.T) {
	frames := []image.Image{picture(10), picture(60), picture(100)}
	if looksStill(frames) {
		t.Error("a block that moves is not a still picture")
	}
}

func TestOneFrameThatDiffersIsEnoughToMove(t *testing.T) {
	frames := []image.Image{picture(20), picture(20), picture(100)}
	if looksStill(frames) {
		t.Error("a video with a title card and then a scene is not still")
	}
}

func TestTooFewFramesSayNothing(t *testing.T) {
	if looksStill(nil) || looksStill([]image.Image{picture(20)}) {
		t.Error("without two frames nothing can be said")
	}
}

func TestFrameDifferenceIsZeroForTheSamePictureAndLargeForADifferentOne(t *testing.T) {
	if d := frameDifference(picture(20), picture(20)); d != 0 {
		t.Errorf("same picture differs by %.2f", d)
	}
	if d := frameDifference(picture(0), picture(100)); d < 20 {
		t.Errorf("moved block differs by only %.2f", d)
	}
}

func serverOf(t *testing.T, bodies map[string][]byte, hits *atomic.Int32) *FramesClient {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		body, ok := bodies[strings.TrimPrefix(r.URL.Path, "/")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)
	return &FramesClient{HTTP: server.Client(), URL: func(id string, n int) string {
		return server.URL + "/" + id + "/" + string(rune('0'+n))
	}}
}

func TestCheckerReadsTheThreeFramesAndRemembersTheAnswer(t *testing.T) {
	still := encoded(t, picture(20))
	bodies := map[string][]byte{"art/1": still, "art/2": still, "art/3": still}
	var hits atomic.Int32
	checker := &Checker{Frames: serverOf(t, bodies, &hits)}

	for i := 0; i < 3; i++ {
		got, err := checker.IsStill("art")
		if err != nil || !got {
			t.Fatalf("round %d: got %v, %v, want a still picture", i, got, err)
		}
	}
	if hits.Load() != 3 {
		t.Errorf("the image server was asked %d times, want 3 (one video, remembered)", hits.Load())
	}
}

func TestCheckerSeesAVideoThatMoves(t *testing.T) {
	bodies := map[string][]byte{
		"clip/1": encoded(t, picture(10)),
		"clip/2": encoded(t, picture(60)),
		"clip/3": encoded(t, picture(100)),
	}
	var hits atomic.Int32
	checker := &Checker{Frames: serverOf(t, bodies, &hits)}

	if got, err := checker.IsStill("clip"); err != nil || got {
		t.Errorf("got %v, %v, want a moving video", got, err)
	}
}

func TestCheckerReportsFramesItCannotRead(t *testing.T) {
	var hits atomic.Int32
	checker := &Checker{Frames: serverOf(t, map[string][]byte{}, &hits)}

	if got, err := checker.IsStill("gone"); err == nil || got {
		t.Errorf("got %v, %v, want an error and no verdict", got, err)
	}
}
