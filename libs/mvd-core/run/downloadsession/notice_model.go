package downloadsession

import "time"

// noticeFor is how long a notice stays.
const noticeFor = 3 * time.Second

// Say gives the person a notice: a short message that stays for a few seconds.
func (s *Session) Say(text string) {
	s.Notice = text
	s.NoticeUntil = time.Now().Add(noticeFor)
}

// NoticeNow is the notice to show at now, and whether there is one.
func (s Session) NoticeNow(now time.Time) (string, bool) {
	if s.Notice != "" && now.Before(s.NoticeUntil) {
		return s.Notice, true
	}

	return "", false
}
