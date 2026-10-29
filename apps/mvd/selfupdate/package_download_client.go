package selfupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// maxPackageBytes is the most a release's zip may be; the real one is about 20 MB.
const maxPackageBytes = 200 << 20

// downloadChecked saves the file at url as dest and checks it against digest (sha256:<hex>, as
// GitHub states it). A file that does not match is removed; nothing is kept that was not checked.
func downloadChecked(ctx context.Context, client *http.Client, url, digest, dest string) error {
	want, ok := strings.CutPrefix(strings.ToLower(strings.TrimSpace(digest)), "sha256:")
	if !ok || len(want) != sha256.Size*2 {
		return errors.New("the release states no checksum for this file")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("the download answered %s", resp.Status)
	}
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	hash := sha256.New()
	n, err := io.Copy(io.MultiWriter(out, hash), io.LimitReader(resp.Body, maxPackageBytes+1))
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	switch {
	case err != nil:
		_ = os.Remove(dest)

		return err
	case n > maxPackageBytes:
		_ = os.Remove(dest)

		return errors.New("the download is larger than any MVD release")
	case hex.EncodeToString(hash.Sum(nil)) != want:
		_ = os.Remove(dest)

		return errors.New("the download does not match its checksum, so it was not used")
	}

	return nil
}
