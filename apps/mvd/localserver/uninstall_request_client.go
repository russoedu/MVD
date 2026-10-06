package localserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"youtube-downloader/apps/mvd/uninstall"
)

// uninstallAnswerTimeout is how long the running app may take: the person is answering
// questions on their screen meanwhile.
const uninstallAnswerTimeout = 15 * time.Minute

// RequestUninstall asks the MVD that is already running on address to remove itself, which
// it does after asking the person on their own screen. That is how an uninstall started
// from Settings > Apps ends the running app instead of leaving its icon behind.
//
// running is false when no MVD answers there, and nothing has been asked. When one does,
// the error is uninstall.ErrDeclined if the person said no, and nil once it has
// accepted and is removing itself.
func RequestUninstall(address string) (running bool, err error) {
	if !isMVD(address) {
		return false, nil
	}

	client := http.Client{Timeout: uninstallAnswerTimeout}
	resp, err := client.Post("http://"+address+"/api/uninstall", "application/json", strings.NewReader("{}"))
	if err != nil {
		return true, err
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusAccepted:
		return true, nil
	case http.StatusConflict:
		return true, uninstall.ErrDeclined
	}

	var body struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body.Error == "" {
		body.Error = resp.Status
	}

	return true, fmt.Errorf("the running MVD refused: %s", body.Error)
}
