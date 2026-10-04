import { retryEntry, retryPlaylist } from '../run'
import type { EntryView, PlaylistView, RunSnapshot } from '../run'
import { formatBytes, formatEta, formatSpeed } from './format-progress.algorithm'

const ACTIVE = new Set(['resolving', 'downloading', 'merging'])

function EntryRow ({ entry }: { entry: EntryView }) {
  const active = ACTIVE.has(entry.state)
  const detail = active
    ? [formatSpeed(entry.speed), formatEta(entry.eta) && `ETA ${formatEta(entry.eta)}`, formatBytes(entry.totalBytes)]
        .filter(Boolean).join(' · ')
    : entry.err

  return (
    <li className={`entry ${entry.state}`}>
      <span className='entry-title'>{entry.title === '' ? entry.videoId : entry.title}</span>
      {entry.channel !== '' && <span className='entry-channel'>{entry.channel}</span>}
      <span className='entry-state'>{entry.official && entry.state === 'done' ? 'official video' : entry.state}</span>
      {active && <progress max={100} value={entry.percent} aria-label={`${entry.title} progress`} />}
      {detail !== '' && <span className='entry-detail'>{detail}</span>}
      {entry.state === 'failed' && (
        <button type='button' onClick={() => void retryEntry(entry.id)}>Retry</button>
      )}
    </li>
  )
}

function PlaylistSection ({ playlist, entries }: { playlist: PlaylistView, entries: EntryView[] }) {
  const heading = playlist.title === '' ? playlist.url : playlist.title

  return (
    <section className='playlist'>
      <header>
        <h2>{heading}</h2>
        <span>
          {playlist.listed ? `${playlist.finished} of ${playlist.total}` : 'reading the list...'}
          {playlist.failed > 0 && ` · ${playlist.failed} failed`}
        </span>
        {playlist.failed > 0 && (
          <button type='button' onClick={() => void retryPlaylist(playlist.index)}>Retry failed</button>
        )}
      </header>
      {playlist.err !== '' && <p className='message error'>{playlist.err}</p>}
      <ul>
        {entries.map(entry => <EntryRow key={entry.id} entry={entry} />)}
      </ul>
    </section>
  )
}

export function QueueView ({ snapshot }: { snapshot: RunSnapshot }) {
  if (snapshot.playlists.length === 0) {
    return <p className='empty'>Nothing queued yet. Paste a link above.</p>
  }
  const byId = new Map(snapshot.entries.map(entry => [entry.id, entry]))

  return (
    <div className='queue'>
      {snapshot.playlists.map(playlist => (
        <PlaylistSection
          key={playlist.index}
          playlist={playlist}
          entries={playlist.entries.flatMap(id => byId.get(id) ?? [])}
        />
      ))}
    </div>
  )
}
