package stillpicture

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// scene paints a 80x45 frame: a grey background with a pattern of bars that depends on
// seed, so that two seeds are two different pictures.
func scene(seed int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 80, 45))
	for y := 0; y < 45; y++ {
		for x := 0; x < 80; x++ {
			v := uint8(70)
			if ((x+seed*7)/6+(y+seed*5)/5)%2 == 0 {
				v = 200
			}
			img.Set(x, y, color.RGBA{R: v, G: v, B: v, A: 255})
		}
	}
	return img
}

// lyric is a still scene with a small caption that changes: a lyric video.
func lyric(line int) *image.RGBA {
	img := scene(0)
	for y := 36; y < 42; y++ {
		for x := 20; x < 60; x++ {
			v := uint8(20)
			if (x/(2+line%3))%2 == 0 {
				v = 240
			}
			img.Set(x, y, color.RGBA{R: v, G: v, B: v, A: 255})
		}
	}
	return img
}

func greys(frames ...image.Image) []greyFrame {
	out := make([]greyFrame, len(frames))
	for i, f := range frames {
		out[i] = greyOf(f)
	}
	return out
}

// repeat builds n frames with a function of the frame's position.
func repeat(n int, build func(i int) image.Image) []greyFrame {
	frames := make([]image.Image, n)
	for i := range frames {
		frames[i] = build(i)
	}
	return greys(frames...)
}

func TestAPictureThatNeverChangesHasNoMotion(t *testing.T) {
	if hasMotion(repeat(30, func(int) image.Image { return scene(0) })) {
		t.Error("the same picture 30 times does not move")
	}
}

func TestAChangingCaptionOverOnePictureHasNoMotion(t *testing.T) {
	if hasMotion(repeat(60, func(i int) image.Image { return lyric(i) })) {
		t.Error("a lyric video over one picture is not a video that moves")
	}
}

func TestADifferentPictureEveryFrameIsMotion(t *testing.T) {
	if !hasMotion(repeat(30, func(i int) image.Image { return scene(i) })) {
		t.Error("a scene that changes all over every frame moves")
	}
}

func TestAFadeIsNotMotion(t *testing.T) {
	frames := repeat(30, func(i int) image.Image {
		img := scene(0)
		for y := 0; y < 45; y++ {
			for x := 0; x < 80; x++ {
				c := img.RGBAAt(x, y)
				k := 1.0 - float64(i)*0.02
				img.Set(x, y, color.RGBA{R: uint8(float64(c.R) * k), G: uint8(float64(c.G) * k), B: uint8(float64(c.B) * k), A: 255})
			}
		}
		return img
	})
	if hasMotion(frames) {
		t.Error("a picture that fades to black is still the same picture")
	}
}

func TestASlowDriftOfOnePixelIsForgiven(t *testing.T) {
	frames := repeat(30, func(i int) image.Image {
		img := image.NewRGBA(image.Rect(0, 0, 80, 45))
		shift := i % 2
		for y := 0; y < 45; y++ {
			for x := 0; x < 80; x++ {
				v := uint8(70)
				if ((x+shift)/6+y/5)%2 == 0 {
					v = 200
				}
				img.Set(x, y, color.RGBA{R: v, G: v, B: v, A: 255})
			}
		}
		return img
	})
	if hasMotion(frames) {
		t.Error("a one pixel shift is registration noise, not motion")
	}
}

func TestTooFewFramesSayNothing(t *testing.T) {
	if hasMotion(nil) || hasMotion(greys(scene(0), scene(3))) {
		t.Error("with fewer than three frames nothing can be said")
	}
}

// sheetOf lays frames out on one storyboard sheet of 10 columns.
func sheetOf(frames []image.Image, width, height, rows, columns int) []byte {
	sheet := image.NewRGBA(image.Rect(0, 0, width*columns, height*rows))
	for i, f := range frames {
		x, y := (i%columns)*width, (i/columns)*height
		for yy := 0; yy < height; yy++ {
			for xx := 0; xx < width; xx++ {
				sheet.Set(x+xx, y+yy, f.At(xx, yy))
			}
		}
	}
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, sheet, &jpeg.Options{Quality: 85})
	return buf.Bytes()
}

func serving(t *testing.T, files map[string][]byte, hits *atomic.Int32) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		body, ok := files[strings.TrimPrefix(r.URL.Path, "/")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)
	return server
}

func storyboardOf(server *httptest.Server, sheets int) Storyboard {
	sb := Storyboard{Width: 80, Height: 45, Rows: 10, Columns: 10}
	for i := 0; i < sheets; i++ {
		sb.Fragments = append(sb.Fragments, StoryboardFragment{URL: fmt.Sprintf("%s/sheet%d.jpg", server.URL, i), DurationSec: 100})
	}
	return sb
}

func TestAStoryboardOfAVideoThatMovesIsNotStatic(t *testing.T) {
	var frames []image.Image
	for i := 0; i < 100; i++ {
		frames = append(frames, scene(i))
	}
	var hits atomic.Int32
	server := serving(t, map[string][]byte{"sheet0.jpg": sheetOf(frames, 80, 45, 10, 10)}, &hits)
	checker := NewChecker(func(string) (Storyboard, error) { return storyboardOf(server, 1), nil })

	static, err := checker.IsStatic("clip")
	if err != nil || static {
		t.Errorf("got static=%v, %v, want a video that moves", static, err)
	}
}

func TestAStoryboardOfALyricVideoIsStaticAndRemembered(t *testing.T) {
	var frames []image.Image
	for i := 0; i < 100; i++ {
		frames = append(frames, lyric(i))
	}
	var hits atomic.Int32
	server := serving(t, map[string][]byte{"sheet0.jpg": sheetOf(frames, 80, 45, 10, 10)}, &hits)
	checker := NewChecker(func(string) (Storyboard, error) { return storyboardOf(server, 1), nil })

	for round := 0; round < 3; round++ {
		static, err := checker.IsStatic("lyrics")
		if err != nil || !static {
			t.Fatalf("round %d: got static=%v, %v, want a static video", round, static, err)
		}
	}
	if hits.Load() != 1 {
		t.Errorf("the sheet was fetched %d times, want once (the answer is remembered)", hits.Load())
	}
}

func TestWithoutAStoryboardTheThreeFramesAreUsed(t *testing.T) {
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, scene(0), &jpeg.Options{Quality: 85})
	still := buf.Bytes()
	var hits atomic.Int32
	server := serving(t, map[string][]byte{"art/1": still, "art/2": still, "art/3": still}, &hits)
	checker := NewChecker(func(string) (Storyboard, error) { return Storyboard{}, fmt.Errorf("no storyboard") })
	checker.FrameURL = func(id string, n int) string { return fmt.Sprintf("%s/%s/%d", server.URL, id, n) }

	static, err := checker.IsStatic("art")
	if err != nil || !static {
		t.Errorf("got static=%v, %v, want a static video", static, err)
	}
}

func TestNothingToReadIsAnError(t *testing.T) {
	var hits atomic.Int32
	server := serving(t, map[string][]byte{}, &hits)
	checker := NewChecker(nil)
	checker.FrameURL = func(id string, n int) string { return fmt.Sprintf("%s/%s/%d", server.URL, id, n) }

	if static, err := checker.IsStatic("gone"); err == nil || static {
		t.Errorf("got static=%v, %v, want an error and no verdict", static, err)
	}
}
