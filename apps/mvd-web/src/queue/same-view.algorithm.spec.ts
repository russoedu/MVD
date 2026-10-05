import type { EntryView, PlaylistView } from '../run'
import { samePlaylist, sameEntry } from './same-view.algorithm'

const entry: EntryView = {
  id:         1,
  playlist:   0,
  index:      0,
  videoId:    'v',
  targetId:   'v',
  title:      'Song',
  channel:    'Band',
  state:      'downloading',
  official:   false,
  err:        '',
  percent:    10,
  downloaded: 0,
  totalBytes: 0,
  speed:      0,
  eta:        -1,
}

const playlist: PlaylistView = {
  index: 0, url: 'u', title: 'Best of', err: '', listed: true, total: 3, finished: 1, failed: 0, active: 1, entries: [1, 2, 3],
}

describe('sameEntry', () => {
  it('matches a copy with the same values, which is what the next snapshot holds', () => {
    expect(sameEntry(entry, { ...entry })).toBe(true)
  })

  it.each([['percent', 11], ['state', 'done'], ['speed', 5], ['err', 'x'], ['title', 'Other']])(
    'sees a change in %s',
    (key, value) => {
      expect(sameEntry(entry, { ...entry, [key]: value })).toBe(false)
    },
  )
})

describe('samePlaylist', () => {
  it('matches a copy, even though its list of ids is a new array', () => {
    expect(samePlaylist(playlist, { ...playlist, entries: [1, 2, 3] })).toBe(true)
  })

  it.each([['finished', 2], ['failed', 1], ['err', 'x'], ['listed', false], ['title', 'Other']])(
    'sees a change in %s',
    (key, value) => {
      expect(samePlaylist(playlist, { ...playlist, [key]: value })).toBe(false)
    },
  )
})
