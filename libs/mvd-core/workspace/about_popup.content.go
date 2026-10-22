package workspace

// projectURL is the page of the project on GitHub, which the About window links to.
const projectURL = "https://github.com/russoedu/MVD"

// aboutParagraphs is what the About window says, the same words as the About window of the
// page around the terminal.
var aboutParagraphs = []string{
	"Remember the 90s? Music videos were an obsession, MTV ran all night, and one crunchy 240p clip on dial-up meant 48 hours of hostage negotiation with the family phone line.",
	"MVD (Music Video Downloader) brings back that visual music. Feed it YouTube video or playlist links and, instead of the lifeless still-image \"Topic\" uploads, it hunts down the official, original music video: first through the link in the video description, and when there is none, by searching for \"<song> <artist> official video\".",
	"Your queue is downloaded concurrently with yt-dlp, with retries, so you can build an offline media wall, fuel a 90s house party, or hoard visual culture before the copyright bots get it.",
	"The look is on purpose: neon green, Barney purple, hard-edged terminal boxes and a bevelled grey button or two.",
}

// aboutFooter is the last line of the About window.
const aboutFooter = "Best viewed with your imagination at 800x600."
