package official

import "testing"

func TestSearchTitle(t *testing.T) {
	cases := map[string]string{
		"Them Bones (2022 Remaster)":                 "Them Bones",
		"Highway Star - Remastered 2012":             "Highway Star",
		"The Power (7\" Version)":                    "The Power",
		"Black Hole Sun - Musora Session":            "Black Hole Sun - Musora Session",
		"Take on Me":                                 "Take on Me",
		"Don't You (Forget About Me) (12\" Version)": "Don't You",
		"(Everything I Do) I Do It for You":          "I Do It for You",
		"[Official] ":                                "[Official]",
		"Walking On Sunshine - 7\" Version":          "Walking On Sunshine",
		"Pump It Up Radio Edit":                      "Pump It Up",
		"Cinema Extended Mix":                        "Cinema",
		"Love Is Gone feat. Jane Doe":                "Love Is Gone",
		"Love Is Gone ft Jane Doe (Radio Version)":   "Love Is Gone",
		"Finally - 7'' Mix 1991":                     "Finally",
		"Last Dance":                                 "Last Dance",
		"Feature Presentation":                       "Feature Presentation",
		"Mix":                                        "Mix",
	}
	for in, want := range cases {
		if got := SearchTitle(in); got != want {
			t.Errorf("SearchTitle(%q) = %q, want %q", in, got, want)
		}
	}
}
