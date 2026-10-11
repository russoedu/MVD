package official

import (
	"youtube-downloader/libs/mvd-core/official/cookiejar"
)

// UseCookies makes every request of the resolver carry the cookies from a
// Netscape cookie file. It returns how many cookies were loaded.
func (r *Resolver) UseCookies(path string) (int, error) {
	jar, n, err := cookiejar.LoadCookieJar(path)
	if err != nil {
		return 0, err
	}
	r.Client.Jar = jar

	return n, nil
}
