package songid

import "testing"

func TestAnAnswerMustBeMentionedByTheUpload(t *testing.T) {
	cases := []struct {
		name     string
		identity Identity
		title    string
		channel  string
		want     bool
	}{
		{"artist and song in the title", Identity{"Le Click", "Tonight Is the Night"}, "Le Click - Tonight Is The Night (1994)", "Pro100 Music", true},
		{"artist in the brackets", Identity{"Le Click", "Tonight Is the Night"}, "Tonight Is The Night (Le Click - Dance Mix )", "La Bouche", true},
		{"artist is the channel", Identity{"Rage", "Run to You"}, "Run to You", "Rage - Topic", true},
		{"accents do not matter", Identity{"Beyoncé", "Halo"}, "BEYONCE - Halo", "x", true},
		{"another artist's song of the same name", Identity{"Madonna", "Vogue"}, "Vogue (Shep Pettibone Mix)", "Some Channel", false},
		{"the artist but another song", Identity{"Le Click", "Tonight Is the Night"}, "Le Click - Call Me", "x", false},
	}
	for _, c := range cases {
		if got := namesTheUpload(c.identity, c.title, c.channel); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
