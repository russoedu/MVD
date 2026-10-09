package official

import "testing"

func TestChannelIsArtist(t *testing.T) {
	yes := []struct{ channel, artist string }{
		{"Nickelback", "Nickelback"},
		{"billyjoelVEVO", "Billy Joel"},
		{"AliceInChainsVEVO", "Alice In Chains"},
		{"INXS Videos", "INXS"},
		{"EarthWindandFireVEVO", "Earth Wind & Fire"},
		{"Earth, Wind & Fire", "Earth Wind & Fire"},
		{"Real McCoy", "Real McCoy"},
		{"OfficialPSY", "PSY"},
		{"The Weeknd", "Weeknd"},
		{"Weeknd", "The Weeknd"},
		{"ADÉLA", "adéla"},
		{"Simon & Garfunkel", "Simon & Garfunkel"},
		{"Michael Sembello (The Master)", "Michael Sembello"},
	}
	for _, c := range yes {
		if !ChannelIsArtist(c.channel, c.artist) {
			t.Errorf("channel %q is the artist %q", c.channel, c.artist)
		}
	}

	no := []struct{ channel, artist string }{
		{"Trance Mission", "ATB"},
		{"Memory Lane", "Bill Withers"},
		{"Sound & Vision", "Simon & Garfunkel"},
		{"Roadrunner Records", "Nickelback"},
		{"Hansonized", "Hanson"},
		{"MusicBoxChannel", "Real McCoy"},
		{"", "ATB"},
		{"ATB", ""},
	}
	for _, c := range no {
		if ChannelIsArtist(c.channel, c.artist) {
			t.Errorf("channel %q is not the artist %q", c.channel, c.artist)
		}
	}
}
