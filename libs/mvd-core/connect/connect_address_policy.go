// Package connect is the opt-in live connection between the running app and the Cuppa
// designer (cuppa ADR 0016): the app publishes the inputs of its pages and applies the
// events the designer sends back. Nothing listens unless the environment asks for it.
package connect

import "strings"

// EnvVar is the environment variable that turns the connection on.
const EnvVar = "MVD_CONNECT"

// DefaultAddress is where the app listens when the variable only says "yes".
const DefaultAddress = "127.0.0.1:7331"

// AddressFrom is the address to listen on for the value of EnvVar, and false when the
// connection was not asked for. A bare yes ("1", "true", "yes", "on") means the default
// address; anything else is taken as the address and checked to be on this computer when
// the app listens.
func AddressFrom(value string) (string, bool) {
	v := strings.TrimSpace(value)
	switch strings.ToLower(v) {
	case "", "0", "false", "no", "off":
		return "", false
	case "1", "true", "yes", "on":
		return DefaultAddress, true
	}
	return v, true
}
