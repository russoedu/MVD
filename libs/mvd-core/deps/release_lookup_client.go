package deps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// AssetInfo describes one build in the latest release of a repository.
type AssetInfo struct {
	// ID changes whenever the asset is uploaded again, which is how a newer build is told
	// from the one already installed.
	ID  int64
	Tag string
	// URL is the address of exactly this build.
	URL string
	// SHA256 is the checksum GitHub publishes for the asset, as hex, or empty when it
	// publishes none.
	SHA256    string
	UpdatedAt time.Time
}

const githubAPI = "https://api.github.com"

// lookupLatest asks GitHub which build of source is the newest.
func lookupLatest(source Source) (AssetInfo, error) {
	return lookupLatestFrom(githubAPI, &http.Client{Timeout: 20 * time.Second}, source)
}

// lookupLatestFrom is lookupLatest against any API address, so a test can answer for GitHub.
func lookupLatestFrom(apiBase string, client *http.Client, source Source) (AssetInfo, error) {
	req, err := http.NewRequest("GET", apiBase+"/repos/"+source.Repo+"/releases/latest", nil)
	if err != nil {
		return AssetInfo{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "mvd")

	resp, err := client.Do(req)
	if err != nil {
		return AssetInfo{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusForbidden, http.StatusTooManyRequests:
		return AssetInfo{}, fmt.Errorf("GitHub is limiting requests for now (HTTP %d)", resp.StatusCode)
	default:
		return AssetInfo{}, fmt.Errorf("asking GitHub about %s: HTTP %d", source.Repo, resp.StatusCode)
	}

	var release struct {
		Tag    string `json:"tag_name"`
		Assets []struct {
			ID        int64     `json:"id"`
			Name      string    `json:"name"`
			URL       string    `json:"browser_download_url"`
			Digest    string    `json:"digest"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return AssetInfo{}, fmt.Errorf("reading GitHub's answer about %s: %w", source.Repo, err)
	}

	for _, asset := range release.Assets {
		if asset.Name != source.Asset {
			continue
		}

		// Only a SHA-256 digest can be checked here; any other kind is ignored.
		checksum := ""
		if hex, ok := strings.CutPrefix(strings.ToLower(asset.Digest), "sha256:"); ok {
			checksum = hex
		}

		return AssetInfo{ID: asset.ID, Tag: release.Tag, URL: asset.URL, SHA256: checksum, UpdatedAt: asset.UpdatedAt}, nil
	}

	return AssetInfo{}, fmt.Errorf("the latest release of %s has no %s", source.Repo, source.Asset)
}
