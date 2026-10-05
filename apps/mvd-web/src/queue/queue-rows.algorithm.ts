import type { EntryView, PlaylistView, RunSnapshot } from '../run'
import { describeEntryDetail, isActive } from './entry-detail.algorithm'

/** Pixel heights. The styles give each row exactly its height, so these decide the layout. */
export const PLAYLIST_ROW_HEIGHT = 52
export const PLAYLIST_ERROR_HEIGHT = 24
export const ENTRY_ROW_HEIGHT = 46
export const ENTRY_LINE_HEIGHT = 22
export const ENTRY_RETRY_HEIGHT = 36

export type QueueRow =
  | { kind: 'playlist', key: string, height: number, playlist: PlaylistView } |
  { kind: 'entry', key: string, height: number, entry: EntryView }

/** An entry grows a line for its progress bar, another for its detail, and one for Retry. */
export function entryRowHeight (entry: EntryView): number {
  return ENTRY_ROW_HEIGHT +
    (isActive(entry) ? ENTRY_LINE_HEIGHT : 0) +
    (describeEntryDetail(entry) === '' ? 0 : ENTRY_LINE_HEIGHT) +
    (entry.state === 'failed' ? ENTRY_RETRY_HEIGHT : 0)
}

/** The whole queue as one flat list: each playlist's heading, then its entries in order. */
export function buildQueueRows (snapshot: RunSnapshot): QueueRow[] {
  const byId = new Map<number, EntryView>()
  for (const entry of snapshot.entries) byId.set(entry.id, entry)

  const rows: QueueRow[] = []
  for (const playlist of snapshot.playlists) {
    rows.push({
      kind:   'playlist',
      key:    `p${playlist.index}`,
      height: PLAYLIST_ROW_HEIGHT + (playlist.err === '' ? 0 : PLAYLIST_ERROR_HEIGHT),
      playlist,
    })
    for (const id of playlist.entries) {
      const entry = byId.get(id)
      if (entry) rows.push({ kind: 'entry', key: `e${entry.id}`, height: entryRowHeight(entry), entry })
    }
  }

  return rows
}
