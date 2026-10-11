package settings

// YesNo is how a switch reads on a settings screen.
func YesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
