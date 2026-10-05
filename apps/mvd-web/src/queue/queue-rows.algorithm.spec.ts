import type { EntryView, RunSnapshot } from '../run'
import {
  buildQueueRows,
  ENTRY_LINE_HEIGHT,
  ENTRY_RETRY_HEIGHT,
  ENTRY_ROW_HEIGHT,
  PLAYLIST_ERROR_HEIGHT,
  PLAYLIST_ROW_HEIGHT,
} from './queue-rows.algorithm'

function entry (over: Partial<EntryView>): EntryView {
  return {
    id:         1,
    playlist:   0,
    index:      0,
    videoId:    'v',
    targetId:   'v',
    title:      'Song',
    channel:    'Band',
    state:      'queued',
    official:   false,
    err:        '',
    percent:    0,
    downloaded: 0,
    totalBytes: 0,
    speed:      0,
    eta:        -1,
    ...over,
  }
}

function snapshot (playlists: Array<{ err?: string, ids: number[] }>, entries: EntryView[]): RunSnapshot {
  return {
    version:   1,
    idle:      false,
    started:   '',
    tally:     { total: entries.length, queued: 0, running: 0, done: 0, official: 0, duplicate: 0, failed: 0, retried: 0 },
    playlists: playlists.map((p, index) => ({
      index,
      url:      `u${index}`,
      title:    `P${index}`,
      err:      p.err ?? '',
      listed:   true,
      total:    p.ids.length,
      finished: 0,
      failed:   0,
      active:   0,
      entries:  p.ids,
    })),
    entries,
  }
}

describe('buildQueueRows', () => {
  it('lists each playlist heading followed by its entries, in the playlist order', () => {
    const rows = buildQueueRows(snapshot(
      [{ ids: [2, 1] }, { ids: [3] }],
      [entry({ id: 1 }), entry({ id: 2 }), entry({ id: 3 })],
    ))

    expect(rows.map(row => row.key)).toEqual(['p0', 'e2', 'e1', 'p1', 'e3'])
    expect(rows.map(row => row.kind)).toEqual(['playlist', 'entry', 'entry', 'playlist', 'entry'])
  })

  it('skips an id that has no entry yet instead of failing', () => {
    const rows = buildQueueRows(snapshot([{ ids: [1, 99] }], [entry({ id: 1 })]))

    expect(rows.map(row => row.key)).toEqual(['p0', 'e1'])
  })

  it('is empty without playlists', () => {
    expect(buildQueueRows(snapshot([], []))).toEqual([])
  })

  it('gives a heading one more line when the playlist failed', () => {
    const rows = buildQueueRows(snapshot([{ err: 'private list', ids: [] }, { ids: [] }], []))

    expect(rows[0].height).toBe(PLAYLIST_ROW_HEIGHT + PLAYLIST_ERROR_HEIGHT)
    expect(rows[1].height).toBe(PLAYLIST_ROW_HEIGHT)
  })

  it('gives an entry a line for its progress bar and a line for its detail', () => {
    const rows = buildQueueRows(snapshot([{ ids: [1, 2, 3, 4] }], [
      entry({ id: 1, state: 'queued' }),
      entry({ id: 2, state: 'failed', err: 'HTTP 403' }),
      entry({ id: 3, state: 'downloading', speed: 1000 }),
      entry({ id: 4, state: 'downloading' }),
    ]))

    expect(rows.slice(1).map(row => row.height)).toEqual([
      ENTRY_ROW_HEIGHT,
      ENTRY_ROW_HEIGHT + ENTRY_LINE_HEIGHT + ENTRY_RETRY_HEIGHT,
      ENTRY_ROW_HEIGHT + 2 * ENTRY_LINE_HEIGHT,
      ENTRY_ROW_HEIGHT + ENTRY_LINE_HEIGHT,
    ])
  })
})
