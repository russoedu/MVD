//go:build !windows

package notification

// showBalloon is false here: only Windows shows its notices without a program.
func showBalloon(Notice) bool { return false }
