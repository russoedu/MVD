import { act, render, screen } from '@testing-library/react'
import type { RunSnapshot } from '../run'
import App from './app'

type Listener = (event: unknown) => void

class FakeEventSource {
  static readonly CLOSED = 2
  static instances: FakeEventSource[] = []
  private readonly listeners = new Map<string, Listener[]>()
  readyState = 1
  closed = false

  constructor (readonly url: string) {
    FakeEventSource.instances.push(this)
  }

  addEventListener (name: string, listener: Listener) {
    this.listeners.set(name, [...(this.listeners.get(name) ?? []), listener])
  }

  close () {
    this.closed = true
  }

  emit (name: string, event: unknown = {}) {
    const listeners = this.listeners.get(name) ?? []
    for (const listener of listeners) listener(event)
  }
}

const snapshot: RunSnapshot = {
  version:   3,
  idle:      false,
  started:   '',
  tally:     { total: 1, queued: 0, running: 1, done: 4, official: 0, duplicate: 0, failed: 2, retried: 0 },
  playlists: [{
    index:    0,
    url:      'https://a.example/p',
    title:    'Best of',
    err:      '',
    listed:   true,
    total:    1,
    finished: 0,
    failed:   0,
    active:   1,
    entries:  [1],
  }],
  entries: [{
    id:         1,
    playlist:   0,
    index:      0,
    videoId:    'v1',
    targetId:   'v1',
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
  }],
}

beforeEach(() => {
  FakeEventSource.instances = []
  Object.assign(globalThis, { EventSource: FakeEventSource })
})

describe('App', () => {
  it('follows the run over server-sent events and draws each snapshot', () => {
    render(<App />)
    const [stream] = FakeEventSource.instances
    expect(stream.url).toBe('/api/events')
    expect(screen.getByRole('alert').textContent).toBe('Connecting...')

    act(() => stream.emit('snapshot', { data: JSON.stringify(snapshot) }))

    expect(screen.queryByRole('alert')).toBeNull()
    expect(screen.getByText('4 done · 1 running · 0 queued · 2 failed')).toBeTruthy()
    expect(screen.getByRole('heading', { name: 'Best of' })).toBeTruthy()
  })

  it('says so when the stream is lost for good', () => {
    render(<App />)
    const [stream] = FakeEventSource.instances

    act(() => {
      stream.readyState = FakeEventSource.CLOSED
      stream.emit('error')
    })

    expect(screen.getByRole('alert').textContent).toContain('Lost contact')
  })

  it('stops listening when it goes away', () => {
    const { unmount } = render(<App />)
    unmount()
    expect(FakeEventSource.instances[0].closed).toBe(true)
  })
})
