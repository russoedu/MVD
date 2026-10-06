package localserver

import (
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// loopbackHosts are the only names this server answers to.
var loopbackHosts = map[string]bool{"127.0.0.1": true, "localhost": true, "::1": true}

// refusal explains why a request is turned away, or is empty when it is let through.
//
// A server on 127.0.0.1 is still reachable by any web page the user opens: the page
// can ask the browser to send a request there. These rules are what stops a page on
// another site from reading the queue or adding URLs.
//
//   - The Host header must be a loopback name. A page on another site can make its
//     own name resolve to 127.0.0.1 (DNS rebinding), and the browser then sends that
//     name as Host, so this is what ends it.
//   - An Origin header, which browsers add to cross-site requests, must itself be a
//     loopback page.
//   - Sec-Fetch-Site, which browsers set and a page cannot forge, must not say
//     cross-site.
//   - A request that changes anything must be JSON. A cross-site page can send a form
//     or text/plain without the browser asking first, but not application/json, which
//     forces a preflight this server never approves.
func refusal(r *http.Request) string {
	if !loopbackHost(r.Host) {
		return "this server only answers to localhost"
	}
	if origin := r.Header.Get("Origin"); origin != "" && !loopbackOrigin(origin) {
		return "requests from other sites are refused"
	}
	if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		return "requests from other sites are refused"
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead && !isJSON(r.Header.Get("Content-Type")) {
		return "send application/json"
	}
	return ""
}

// loopbackHost reports whether a Host header, with or without a port, names this machine.
func loopbackHost(hostport string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	return loopbackHosts[strings.Trim(strings.ToLower(host), "[]")]
}

// loopbackOrigin reports whether an Origin header is a page served from this machine.
func loopbackOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return false
	}
	return loopbackHosts[strings.Trim(strings.ToLower(parsed.Hostname()), "[]")]
}

// isJSON reports whether a Content-Type is application/json, parameters aside.
func isJSON(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	return err == nil && mediaType == "application/json"
}
