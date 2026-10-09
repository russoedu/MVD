// Package stillpicture tells a video that really moves from one that is only a
// picture with a song over it: an art track, a cover with the audio, a lyric video
// whose background never changes, a visualizer.
//
// It looks at the storyboard YouTube keeps of every video: one small frame every
// second or two, laid out in a few sheets, which is what the player shows when the
// seek bar is hovered. Each frame is divided by its own local background so that a
// fade or a change of light takes no part, a pixel is forgiven a step of a pixel in
// every direction, and what is measured is the share of the frame that changed. A
// real video changes a fifth to two fifths of its frame between one frame and the
// next; a still changes under two per cent, a lyric video over one picture about
// five. The idea comes from scanmate's pixel comparison, which does the same to find
// ink that a scan added to a document.
package stillpicture
