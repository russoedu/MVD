import { memo } from 'react'
import { retryEntry } from '../run'
import type { EntryView } from '../run'
import { describeEntryDetail, isActive } from './entry-detail.algorithm'
import { sameEntry } from './same-view.algorithm'

function EntryRowContent ({ entry }: { entry: EntryView }) {
  const active = isActive(entry)
  const detail = describeEntryDetail(entry)

  return (
    <div className={`entry ${entry.state}`}>
      <span className='entry-title'>{entry.title === '' ? entry.videoId : entry.title}</span>
      {entry.channel !== '' && <span className='entry-channel'>{entry.channel}</span>}
      <span className='entry-state'>{entry.official && entry.state === 'done' ? 'official video' : entry.state}</span>
      {active && <progress max={100} value={entry.percent} aria-label={`${entry.title} progress`} />}
      {detail !== '' && <span className='entry-detail'>{detail}</span>}
      {entry.state === 'failed' && (
        <button type='button' onClick={() => void retryEntry(entry.id)}>Retry</button>
      )}
    </div>
  )
}

/**
 * Each snapshot arrives as new objects, so the entry is compared by value: a progress tick
 * for one download redraws that row and leaves the other mounted rows alone.
 */
export const EntryRow = memo(EntryRowContent, (a, b) => sameEntry(a.entry, b.entry))
