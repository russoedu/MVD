package official

import "regexp"

// advertised are the words an upload puts in its title when it is proud of its picture.
var advertised = regexp.MustCompile(`(?i)\b(1080p|1440p|2160p|720p|4k|uhd|hd|high quality|hq|remastered)\b`)

// AdvertisesQuality reports whether a title says the upload is in high quality ("(HD,
// 1080p, 16:9)"). It is only a hint for which uploads have their formats looked up
// first; the formats themselves decide.
func AdvertisesQuality(title string) bool {
	return advertised.MatchString(title)
}
