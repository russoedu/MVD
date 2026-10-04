import { render, screen } from '@testing-library/react'
import type { EntryView, RunSnapshot } from '../run'
import { QueueView } from './queue-view.component'

function entry (over: Partial<EntryView>): EntryView {
  return {
    id:         1,
    playlist:   0,
    index:      0,
    videoId:    'v1',
    targetId:   'v1',
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

function snapshot (entries: EntryView[], over: Partial<RunSnapshot['playlists'][0]> = {}): RunSnapshot {
  return {
    version:   1,
    idle:      false,
    started:   '',
    tally:     { total: entries.length, queued: 0, running: 0, done: 0, official: 0, duplicate: 0, failed: 0, retried: 0 },
    playlists: [{
      index:    0,
      url:      'https://a.example/p',
      title:    'Best of',
      err:      '',
      listed:   true,
      total:    entries.length,
      finished: 0,
      failed:   0,
      active:   0,
      entries:  entries.map(e => e.id),
      ...over,
    }],
    entries,
  }
}

describe('QueueView', () => {
  it('invites a first link when nothing is queued', () => {
    render(<QueueView snapshot={{ ...snapshot([]), playlists: [] }} />)
    expect(screen.getByText(/Nothing queued yet/)).toBeTruthy()
  })

  it('shows a playlist with its entries and progress', () => {
    const running = entry({ id: 1, title: 'Running', state: 'downloading', percent: 40, speed: 2 * 1024 * 1024, eta: 65, totalBytes: 50 * 1024 * 1024 })
    const waiting = entry({ id: 2, title: 'Waiting', state: 'queued' })
    render(<QueueView snapshot={snapshot([running, waiting])} />)

    expect(screen.getByRole('heading', { name: 'Best of' })).toBeTruthy()
    expect(screen.getByLabelText<HTMLProgressElement>('Running progress').value).toBe(40)
    expect(screen.getByText('2.0 MB/s · ETA 1:05 · 50 MB')).toBeTruthy()
    expect(screen.getByText('Waiting')).toBeTruthy()
  })

  it('offers a retry only for failed entries and shows why they failed', () => {
    const fine = entry({ id: 1, title: 'Fine', state: 'done' })
    const broken = entry({ id: 2, title: 'Broken', state: 'failed', err: 'HTTP 403' })
    render(<QueueView snapshot={snapshot([fine, broken], { failed: 1, finished: 2 })} />)

    expect(screen.getByText('HTTP 403')).toBeTruthy()
    expect(screen.getAllByRole('button', { name: 'Retry' })).toHaveLength(1)
    expect(screen.getByRole('button', { name: 'Retry failed' })).toBeTruthy()
  })

  it('says it is still reading a playlist that has not been listed', () => {
    render(<QueueView snapshot={snapshot([], { listed: false, title: '' })} />)
    expect(screen.getByText('reading the list...')).toBeTruthy()
  })
})
