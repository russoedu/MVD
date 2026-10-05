import { memo } from 'react'
import { retryPlaylist } from '../run'
import type { PlaylistView } from '../run'
import { samePlaylist } from './same-view.algorithm'

function PlaylistHeaderContent ({ playlist }: { playlist: PlaylistView }) {
  const heading = playlist.title === '' ? playlist.url : playlist.title

  return (
    <header className='playlist-header'>
      <h2>{heading}</h2>
      <span>
        {playlist.listed ? `${playlist.finished} of ${playlist.total}` : 'reading the list...'}
        {playlist.failed > 0 && ` · ${playlist.failed} failed`}
      </span>
      {playlist.failed > 0 && (
        <button type='button' onClick={() => void retryPlaylist(playlist.index)}>Retry failed</button>
      )}
      {playlist.err !== '' && <p className='message error playlist-error' title={playlist.err}>{playlist.err}</p>}
    </header>
  )
}

export const PlaylistHeaderRow = memo(PlaylistHeaderContent, (a, b) => samePlaylist(a.playlist, b.playlist))
