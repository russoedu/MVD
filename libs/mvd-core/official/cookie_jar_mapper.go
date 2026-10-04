package official

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// LoadCookieJar reads a Netscape cookie file (the format browsers export
// and yt-dlp writes) into a cookie jar usable by an http.Client.
func LoadCookieJar(path string) (http.CookieJar, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = f.Close() }()

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, 0, err
	}

	count := 0
	byHost := map[string][]*http.Cookie{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		httpOnly := false
		if strings.HasPrefix(line, "#HttpOnly_") {
			httpOnly = true
			line = strings.TrimPrefix(line, "#HttpOnly_")
		} else if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			continue
		}
		domain := strings.TrimPrefix(fields[0], ".")
		expires, _ := strconv.ParseInt(fields[4], 10, 64)
		c := &http.Cookie{
			Name:     fields[5],
			Value:    fields[6],
			Path:     fields[2],
			Domain:   domain,
			Secure:   strings.EqualFold(fields[3], "TRUE"),
			HttpOnly: httpOnly,
		}
		if expires > 0 {
			c.Expires = time.Unix(expires, 0)
		}
		byHost[domain] = append(byHost[domain], c)
		count++
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}
	for host, cookies := range byHost {
		u := &url.URL{Scheme: "https", Host: host, Path: "/"}
		jar.SetCookies(u, cookies)
	}
	if count == 0 {
		return nil, 0, fmt.Errorf("no cookies found in %s", path)
	}
	return jar, count, nil
}

// UseCookies makes every request of the resolver carry the cookies from a
// Netscape cookie file. It returns how many cookies were loaded.
func (r *Resolver) UseCookies(path string) (int, error) {
	jar, n, err := LoadCookieJar(path)
	if err != nil {
		return 0, err
	}
	r.Client.Jar = jar
	return n, nil
}
