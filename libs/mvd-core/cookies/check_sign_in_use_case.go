package cookies

import (
	"context"
	"os"

	"youtube-downloader/libs/mvd-core/ytdlp"
)

// CheckSignIn reads the cookies of the browser spec and says whether they carry a YouTube login,
// and whether that login is Premium. The sign-in window must be closed, or its cookie database
// is locked and nothing can be read.
func CheckSignIn(ctx context.Context, bin, spec string, extraArgs []string) (signedIn, premium bool) {
	dir, err := os.MkdirTemp("", "mvd-sign-in")
	if err != nil {
		return false, false
	}
	defer func() { _ = os.RemoveAll(dir) }()
	file := dir + string(os.PathSeparator) + "cookies.txt"
	if err := ytdlp.ExportCookies(ctx, bin, spec, file, premiumProbeURL, extraArgs); err != nil || !hasYouTubeSession(file) {
		return false, false
	}

	return true, ytdlp.HasPremiumAudio(ctx, bin, file, premiumProbeURL, extraArgs)
}
