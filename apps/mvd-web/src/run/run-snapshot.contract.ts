// The run as the server sends it (libs/mvd-server/snapshot).

export type EntryState =
  | 'queued' |
  'resolving' |
  'downloading' |
  'merging' |
  'done' |
  'duplicate' |
  'failed' |
  'unknown'

export interface Tally {
  total:     number
  queued:    number
  running:   number
  done:      number
  official:  number
  duplicate: number
  failed:    number
  retried:   number
}

export interface PlaylistView {
  index:    number
  url:      string
  title:    string
  err:      string
  listed:   boolean
  total:    number
  finished: number
  failed:   number
  active:   number
  entries:  number[]
}

export interface EntryView {
  id:         number
  playlist:   number
  index:      number
  videoId:    string
  targetId:   string
  title:      string
  channel:    string
  state:      EntryState
  official:   boolean
  err:        string
  percent:    number
  downloaded: number
  totalBytes: number
  /** Bytes per second, 0 when unknown. */
  speed:      number
  /** Seconds, -1 when unknown. */
  eta:        number
}

export interface RunSnapshot {
  version:   number
  idle:      boolean
  started:   string
  tally:     Tally
  playlists: PlaylistView[]
  entries:   EntryView[]
}

/** What became of each line sent to POST /api/sources. */
export interface AddSourcesResult {
  added:      string[]
  duplicates: string[]
  rejected:   string[]
}
