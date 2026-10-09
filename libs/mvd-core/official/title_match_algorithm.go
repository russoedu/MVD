package official

import (
	"regexp"
	"strings"
)

// titleNoise are the words a title adds around a song's name without changing
// which song it is: "Them Bones (2022 Remaster)" is "Them Bones", an
// "(Official HD Video)" is the song's video.
var titleNoise = map[string]bool{
	"official": true, "video": true, "music": true, "hd": true, "4k": true, "uhd": true, "hq": true,
	"1080p": true, "remaster": true, "remastered": true, "version": true, "edit": true, "radio": true,
	"single": true, "album": true, "mono": true, "stereo": true, "original": true, "explicit": true,
	"clean": true, "feat": true, "ft": true, "featuring": true,
}

var yearToken = regexp.MustCompile(`^(19|20)\d\d$`)

// titleTokens lowercases a title and splits it into words, leaving out the
// bracketed parts, the noise words and years ("Killer 2000" is "Killer"), and
// the words of the artists named in skip, unless keep has them (a song called
// "Fire" by "Earth, Wind & Fire" keeps its word).
func titleTokens(title string, skip, keep map[string]bool) []string {
	var tokens []string
	for _, word := range strings.Fields(normalize(bracketed.ReplaceAllString(title, " "))) {
		if titleNoise[word] || yearToken.MatchString(word) || (skip[word] && !keep[word]) {
			continue
		}
		tokens = append(tokens, word)
	}
	return tokens
}

func tokenSet(tokens []string) map[string]bool {
	set := make(map[string]bool, len(tokens))
	for _, token := range tokens {
		set[token] = true
	}
	return set
}

// TitleSimilarity says how much a video's title names the song, from 0 (not
// the song) to 1 (exactly it). The video's title usually adds the artist ("ATB
// - Killer"), words like "Official Video" and sometimes a year, none of which
// count against it; spacing differences ("Run Away" and "Runaway") do not
// either. A title that has only some of the song's words ("Alive" for "Alive
// And Kicking") scores low.
func TitleSimilarity(song, candidate string, artists []string) float64 {
	artistWords := map[string]bool{}
	for _, artist := range artists {
		for _, word := range strings.Fields(normalize(artist)) {
			artistWords[word] = true
		}
	}

	songTokens := titleTokens(song, nil, nil)
	candidateTokens := titleTokens(candidate, artistWords, tokenSet(songTokens))
	if len(songTokens) == 0 || len(candidateTokens) == 0 {
		return 0
	}

	songKey, candidateKey := strings.Join(songTokens, ""), strings.Join(candidateTokens, "")
	if songKey == candidateKey {
		return 1
	}

	shared := 0
	candidateSet := tokenSet(candidateTokens)
	for token := range tokenSet(songTokens) {
		if candidateSet[token] {
			shared++
		}
	}
	dice := 2 * float64(shared) / float64(len(tokenSet(songTokens))+len(candidateSet))

	// A spelling slip or a split word ("Mr Brightside", "Runaway") is close enough.
	if ratio := editSimilarity(songKey, candidateKey); ratio >= 0.85 && ratio*0.95 > dice {
		return ratio * 0.95
	}
	return dice
}

// editSimilarity is 1 minus the edit distance over the length of the longer text.
func editSimilarity(a, b string) float64 {
	longest := max(len([]rune(a)), len([]rune(b)))
	if longest == 0 {
		return 1
	}
	return 1 - float64(editDistance([]rune(a), []rune(b)))/float64(longest)
}

func editDistance(a, b []rune) int {
	previous := make([]int, len(b)+1)
	for j := range previous {
		previous[j] = j
	}
	for i := 1; i <= len(a); i++ {
		current := make([]int, len(b)+1)
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			current[j] = min(previous[j]+1, current[j-1]+1, previous[j-1]+cost)
		}
		previous = current
	}
	return previous[len(b)]
}
