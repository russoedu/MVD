package selfupdate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// releasesURL lists the releases of the project, newest first. The repository also publishes
// the terminal app (mvd-tui@x.y.z), so "the latest release" would not always be this one.
const releasesURL = "https://api.github.com/repos/russoedu/MVD/releases?per_page=30"

const tagPrefix = "mvd@"

type releaseJSON struct {
	Tag        string `json:"tag_name"`
	Page       string `json:"html_url"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name   string `json:"name"`
		URL    string `json:"browser_download_url"`
		Digest string `json:"digest"`
	} `json:"assets"`
}

// newestRelease asks the list at url for the newest published MVD.
func newestRelease(ctx context.Context, client *http.Client, url string) (Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("GitHub answered %s", resp.Status)
	}
	var list []releaseJSON
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return Release{}, fmt.Errorf("cannot read the list of releases: %w", err)
	}
	for _, r := range list {
		if r.Draft || r.Prerelease || !strings.HasPrefix(r.Tag, tagPrefix) {
			continue
		}
		out := Release{Version: strings.TrimPrefix(r.Tag, tagPrefix), Page: r.Page}
		for _, a := range r.Assets {
			out.Assets = append(out.Assets, Asset{Name: a.Name, URL: a.URL, Digest: a.Digest})
		}

		return out, nil
	}

	return Release{}, errors.New("no MVD release found")
}
