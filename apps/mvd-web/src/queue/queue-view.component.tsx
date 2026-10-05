import { useMemo } from 'react'
import type { RunSnapshot } from '../run'
import { EntryRow } from './entry-row.component'
import { PlaylistHeaderRow } from './playlist-header-row.component'
import { buildQueueRows } from './queue-rows.algorithm'
import type { QueueRow } from './queue-rows.algorithm'
import { WindowedRows } from './windowed-rows.component'

function renderQueueRow (row: QueueRow) {
  return row.kind === 'playlist'
    ? <PlaylistHeaderRow playlist={row.playlist} />
    : <EntryRow entry={row.entry} />
}

export function QueueView ({ snapshot }: { snapshot: RunSnapshot }) {
  const rows = useMemo(() => buildQueueRows(snapshot), [snapshot])

  if (snapshot.playlists.length === 0) {
    return <p className='empty'>Nothing queued yet. Paste a link above.</p>
  }

  return <WindowedRows className='queue' rows={rows} renderRow={renderQueueRow} />
}
