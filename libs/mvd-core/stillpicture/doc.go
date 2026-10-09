// Package stillpicture tells a video that is only a picture with a song over it (an
// art track, an "audio" upload with the cover) from one that moves. YouTube serves
// three frames of every video, at 25, 50 and 75 percent of its length; if they are
// the same picture, nothing in the video moves. It looks at those, so the video itself
// is never downloaded.
package stillpicture
