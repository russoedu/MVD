package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// userAgent is what the catalogues are told the requests come from.
const userAgent = "Mozilla/5.0 (compatible; MVD music search)"

// maxRetries is how many times a catalogue that asks to be asked less often is asked again.
const maxRetries = 2

// retryDelay is how long a catalogue is left alone when it does not say how long (a variable, so a
// test need not wait).
var retryDelay = 2 * time.Second

// getJSON asks url and reads the answer as JSON into out. A catalogue that says it is asked too often
// (429, or 403 for iTunes) is waited for, as long as it asks or a couple of seconds, and asked again.
func getJSON(ctx context.Context, client *http.Client, url string, out any) error {
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", userAgent)
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		body, readErr := io.ReadAll(io.LimitReader(res.Body, 8<<20))
		_ = res.Body.Close()
		if readErr != nil {
			return readErr
		}
		if (res.StatusCode == http.StatusTooManyRequests || res.StatusCode == http.StatusForbidden) && attempt < maxRetries {
			if !sleep(ctx, retryWait(res)) {
				return ctx.Err()
			}

			continue
		}
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("answered %s", res.Status)
		}

		return json.Unmarshal(body, out)
	}
}

// retryWait is how long a catalogue asked to be left alone, or the usual wait.
func retryWait(res *http.Response) time.Duration {
	if seconds, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil && seconds >= 0 && seconds <= 30 {
		return time.Duration(seconds) * time.Second
	}

	return retryDelay
}

// sleep waits for d, and says false when ctx ended first.
func sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
