package official

import (
	"fmt"
	"net/http"
	"os"
)

// LoadCookieJar reads the Netscape cookie file at path into a cookie jar usable by an
// http.Client. A file without cookies is an error.
func LoadCookieJar(path string) (http.CookieJar, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = f.Close() }()

	jar, count, err := cookieJarFromNetscape(f)
	if err != nil {
		return nil, 0, err
	}
	if count == 0 {
		return nil, 0, fmt.Errorf("no cookies found in %s", path)
	}

	return jar, count, nil
}
