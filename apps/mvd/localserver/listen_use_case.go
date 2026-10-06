package localserver

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"syscall"
	"time"
)

// DefaultAddress is where the app lives when nothing says otherwise. Loopback only:
// the downloads belong to this user on this machine, and the API can queue work.
const DefaultAddress = "127.0.0.1:8421"

// Listen opens the address. If it is taken it says whether the occupant is another
// MVD (so the caller can open that one instead of starting a second), and if it is
// something else it takes any free port on the same host.
func Listen(address string) (listener net.Listener, runningHere string, err error) {
	listener, err = net.Listen("tcp", address)
	if err == nil {
		return listener, "", nil
	}
	if !errors.Is(err, syscall.EADDRINUSE) && !isAddressInUse(err) {
		return nil, "", err
	}
	if isMVD(address) {
		return nil, address, nil
	}

	host, _, splitErr := net.SplitHostPort(address)
	if splitErr != nil {
		return nil, "", splitErr
	}
	listener, err = net.Listen("tcp", net.JoinHostPort(host, "0"))
	return listener, "", err
}

// isAddressInUse recognises "address already in use" where the error is not
// wrapping a syscall errno, which is how Windows reports it.
func isAddressInUse(err error) bool {
	var opErr *net.OpError
	if !errors.As(err, &opErr) {
		return false
	}
	var errno syscall.Errno
	if errors.As(opErr.Err, &errno) {
		// WSAEADDRINUSE
		return errno == 10048
	}
	return false
}

// isMVD reports whether an MVD is answering on address: this version answers its ping
// route, and versions up to 0.0.23 return a snapshot from the state route, which nothing
// else on the port would.
func isMVD(address string) bool {
	client := http.Client{Timeout: 700 * time.Millisecond}
	if answersPing(client, address) {
		return true
	}
	resp, err := client.Get("http://" + address + "/api/state")
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	var body struct {
		Version *int64          `json:"version"`
		Tally   json.RawMessage `json:"tally"`
	}
	return json.NewDecoder(resp.Body).Decode(&body) == nil && body.Version != nil && body.Tally != nil
}

// answersPing reports whether address answers the ping route the way MVD does.
func answersPing(client http.Client, address string) bool {
	resp, err := client.Get("http://" + address + "/api/ping")
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	var body struct {
		App string `json:"app"`
	}
	return resp.StatusCode == http.StatusOK && json.NewDecoder(resp.Body).Decode(&body) == nil && body.App == "mvd"
}
