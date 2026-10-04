package official

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// do sends a request with browser-like headers, retrying on 429 and 5xx.
func (r *Resolver) do(req *http.Request) ([]byte, int, error) {
	req.Header.Set("User-Agent", r.UserAgent)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	if r.Client.Jar == nil {
		// Skip the EU cookie consent interstitial.
		req.Header.Set("Cookie", "SOCS=CAI; CONSENT=YES+cb")
	}

	attempts := r.Attempts
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(time.Duration(1<<uint(i)) * time.Second)
		}
		resp, err := r.Client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("HTTP status %s", resp.Status)
			continue
		}
		return body, resp.StatusCode, nil
	}
	return nil, 0, lastErr
}

func (r *Resolver) get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	body, status, err := r.do(req)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("HTTP status %d", status)
	}
	return body, nil
}

// fetchNext calls the innertube "next" endpoint for a video. The client
// context is taken from the watch page when available so the request looks
// like the one the page itself makes.
func (r *Resolver) fetchNext(videoID string, pageHTML []byte) ([]byte, error) {
	ctx := map[string]interface{}{
		"client": map[string]interface{}{
			"clientName":    "WEB",
			"clientVersion": "2.20240101.00.00",
			"hl":            "en",
			"gl":            "US",
		},
	}
	clientVersion := "2.20240101.00.00"
	if pageHTML != nil {
		if raw := extractJSONAfter(pageHTML, `"INNERTUBE_CONTEXT":`); raw != nil {
			var pageCtx map[string]interface{}
			if json.Unmarshal(raw, &pageCtx) == nil {
				ctx = pageCtx
				if c, ok := pageCtx["client"].(map[string]interface{}); ok {
					if v, ok := c["clientVersion"].(string); ok && v != "" {
						clientVersion = v
					}
				}
			}
		}
	}

	payload, err := json.Marshal(map[string]interface{}{
		"context": ctx,
		"videoId": videoID,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", r.NextURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://www.youtube.com")
	req.Header.Set("Referer", r.WatchBase+videoID)
	req.Header.Set("X-YouTube-Client-Name", "1")
	req.Header.Set("X-YouTube-Client-Version", clientVersion)

	body, status, err := r.do(req)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("HTTP status %d", status)
	}
	return body, nil
}

// lookupAuthor asks the oEmbed endpoint who uploaded a video.
// ok=false means the video does not exist (404).
func (r *Resolver) lookupAuthor(videoID string) (author string, ok bool, err error) {
	body, err := func() ([]byte, error) {
		req, err := http.NewRequest("GET", r.OEmbedURL+"https://www.youtube.com/watch?v="+videoID, nil)
		if err != nil {
			return nil, err
		}
		body, status, err := r.do(req)
		if err != nil {
			return nil, err
		}
		if status == http.StatusNotFound {
			return nil, nil
		}
		if status != http.StatusOK {
			return nil, fmt.Errorf("HTTP status %d", status)
		}
		return body, nil
	}()
	if err != nil {
		return "", false, err
	}
	if body == nil {
		return "", false, nil
	}
	var o struct {
		AuthorName string `json:"author_name"`
	}
	if err := json.Unmarshal(body, &o); err != nil {
		return "", false, err
	}
	return o.AuthorName, true, nil
}
