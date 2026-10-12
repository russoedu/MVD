package connect

import "testing"

func TestAddressFrom(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"", "", false},
		{"0", "", false},
		{"off", "", false},
		{"1", DefaultAddress, true},
		{"True", DefaultAddress, true},
		{" 127.0.0.1:9000 ", "127.0.0.1:9000", true},
	}
	for _, c := range cases {
		got, ok := AddressFrom(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("AddressFrom(%q) = %q, %v; want %q, %v", c.in, got, ok, c.want, c.ok)
		}
	}
}
