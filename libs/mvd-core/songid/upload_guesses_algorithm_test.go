package songid

import (
	"reflect"
	"testing"
)

func TestGuessesReadTheArtistFromTheBrackets(t *testing.T) {
	got := guessesFrom("Tonight Is The Night (Le Click - Dance Mix )", "La Bouche")
	want := []Identity{
		{Artist: "Le Click", Title: "Tonight Is The Night"},
		{Artist: "La Bouche", Title: "Tonight Is The Night"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestGuessesTryBothHalvesOfADashedTitle(t *testing.T) {
	got := guessesFrom("Masterboy - Pump It Up (Radio Edit) (1991)", "Some Channel")
	want := []Identity{
		{Artist: "Masterboy", Title: "Pump It Up"},
		{Artist: "Pump It Up", Title: "Masterboy"},
		{Artist: "Some Channel", Title: "Masterboy - Pump It Up"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestGuessesUseTheChannelOfAnArtTrack(t *testing.T) {
	got := guessesFrom("Run to You", "Rage - Topic")
	want := []Identity{{Artist: "Rage", Title: "Run to You"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestGuessesOfNothingAreNone(t *testing.T) {
	if got := guessesFrom("", ""); len(got) != 0 {
		t.Errorf("got %+v", got)
	}
}
