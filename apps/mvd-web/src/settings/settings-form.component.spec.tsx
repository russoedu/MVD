import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import type { SettingsDocument } from './settings.contract'
import { SettingsForm } from './settings-form.component'

const fetchMock = jest.fn()

const document_: SettingsDocument = {
  settings: {
    outputDir:                  '/music',
    videoQuality:               'best',
    audioQuality:               'best',
    mergeOutputFormat:          'mp4',
    outputTemplate:             '%(title)s.%(ext)s',
    maxConcurrentDownloads:     4,
    concurrentFragments:        4,
    downloadOfficialMusicVideo: false,
    autoRetry:                  true,
    cookies:                    'all',
    createLogFile:              true,
    logDir:                     '/logs',
  },
  options: {
    videoQualities: ['best', '1080p'], audioQualities: ['best', 'low'], mergeFormats: ['mp4', 'mkv'], browsers: ['firefox'],
  },
}

type Reply = { status: number, body: unknown }

function serve (routes: Record<string, Reply | (() => Reply)>) {
  fetchMock.mockImplementation((path: string, init: RequestInit = {}) => {
    const route = routes[`${init.method ?? 'GET'} ${path}`]
    if (!route) return Promise.reject(new Error(`unexpected ${init.method} ${path}`))
    const { status, body } = typeof route === 'function' ? route() : route

    return Promise.resolve({ ok: status < 400, status, json: () => Promise.resolve(body) })
  })
}

const input = (name: string) => screen.getByLabelText<HTMLInputElement>(name)
const bodyOf = (call: unknown[]) => JSON.parse((call[1] as RequestInit).body as string) as Record<string, unknown>

beforeEach(() => {
  fetchMock.mockReset()
  Object.defineProperty(globalThis, 'fetch', { value: fetchMock, configurable: true, writable: true })
})

describe('SettingsForm', () => {
  it('shows the stored values and offers the server\'s choices', async () => {
    serve({ 'GET /api/settings': { status: 200, body: document_ } })
    render(<SettingsForm runStarted={false} />)

    await waitFor(() => expect(input('Download folder').value).toBe('/music'))
    expect(screen.getByLabelText<HTMLSelectElement>('Video quality').value).toBe('best')
    expect([...screen.getByLabelText<HTMLSelectElement>('Video quality').options].map(o => o.value)).toEqual(['best', '1080p'])
    expect([...screen.getByLabelText<HTMLSelectElement>('Browser sign-in (cookies)').options].map(o => o.value)).toEqual(['all', 'off', 'firefox'])
    expect(input('Videos at once').value).toBe('4')
  })

  it('says so when the settings cannot be loaded', async () => {
    serve({ 'GET /api/settings': { status: 500, body: { error: 'cannot read the settings: broken' } } })
    render(<SettingsForm runStarted={false} />)

    await waitFor(() => expect(screen.getByRole('status').textContent).toContain('cannot read the settings'))
  })

  it('saves what was changed and confirms', async () => {
    const saved = { ...document_, settings: { ...document_.settings, videoQuality: '1080p', maxConcurrentDownloads: 2 } }
    serve({ 'GET /api/settings': { status: 200, body: document_ }, 'PUT /api/settings': { status: 200, body: saved } })
    render(<SettingsForm runStarted={false} />)
    await waitFor(() => expect(input('Download folder').value).toBe('/music'))

    fireEvent.change(screen.getByLabelText('Video quality'), { target: { value: '1080p' } })
    fireEvent.change(input('Videos at once'), { target: { value: '2' } })
    fireEvent.click(screen.getByRole('button', { name: 'Save settings' }))

    await waitFor(() => expect(screen.getByRole('status').textContent).toBe('Saved.'))
    const put = fetchMock.mock.calls.find(call => (call[1] as RequestInit | undefined)?.method === 'PUT') as unknown[]
    expect(bodyOf(put)).toMatchObject({ videoQuality: '1080p', maxConcurrentDownloads: 2, outputDir: '/music', cookies: 'all' })
  })

  it('tells the person to restart when downloads have already started', async () => {
    serve({ 'GET /api/settings': { status: 200, body: document_ }, 'PUT /api/settings': { status: 200, body: document_ } })
    render(<SettingsForm runStarted />)
    await waitFor(() => expect(input('Download folder').value).toBe('/music'))

    fireEvent.click(screen.getByRole('button', { name: 'Save settings' }))

    await waitFor(() => expect(screen.getByRole('status').textContent).toContain('Restart MVD'))
  })

  it('marks the field the server rejected and keeps what was typed', async () => {
    serve({
      'GET /api/settings': { status: 200, body: document_ },
      'PUT /api/settings': { status: 400, body: { error: 'the settings are not valid', fields: { maxConcurrentDownloads: 'enter a number from 1 to 32' } } },
    })
    render(<SettingsForm runStarted={false} />)
    await waitFor(() => expect(input('Download folder').value).toBe('/music'))

    fireEvent.change(input('Videos at once'), { target: { value: '99' } })
    fireEvent.click(screen.getByRole('button', { name: 'Save settings' }))

    await waitFor(() => expect(screen.getByText('enter a number from 1 to 32')).toBeTruthy())
    expect(input('Videos at once').value).toBe('99')
    expect(screen.getByRole('status').textContent).toBe('the settings are not valid')
  })

  it('fills the folder from the chooser, starting it at the current folder', async () => {
    serve({
      'GET /api/settings':      { status: 200, body: document_ },
      'POST /api/folders/pick': { status: 200, body: { path: '/chosen', cancelled: false } },
    })
    render(<SettingsForm runStarted={false} />)
    await waitFor(() => expect(input('Download folder').value).toBe('/music'))

    fireEvent.click(screen.getAllByRole('button', { name: 'Browse...' })[0])

    await waitFor(() => expect(input('Download folder').value).toBe('/chosen'))
    const pick = fetchMock.mock.calls.find(call => (call[1] as RequestInit | undefined)?.method === 'POST') as unknown[]
    expect(bodyOf(pick)).toEqual({ start: '/music' })
  })

  it('leaves the folder alone when the chooser is cancelled', async () => {
    serve({
      'GET /api/settings':      { status: 200, body: document_ },
      'POST /api/folders/pick': { status: 200, body: { path: '', cancelled: true } },
    })
    render(<SettingsForm runStarted={false} />)
    await waitFor(() => expect(input('Download folder').value).toBe('/music'))

    fireEvent.click(screen.getAllByRole('button', { name: 'Browse...' })[0])

    await waitFor(() => expect(screen.getAllByRole('button', { name: 'Browse...' })[0]).toBeTruthy())
    expect(input('Download folder').value).toBe('/music')
  })

  it('explains when the machine has no chooser, and the folder can still be typed', async () => {
    serve({
      'GET /api/settings':      { status: 200, body: document_ },
      'POST /api/folders/pick': { status: 501, body: { error: 'no folder chooser is available here; type the path instead' } },
    })
    render(<SettingsForm runStarted={false} />)
    await waitFor(() => expect(input('Download folder').value).toBe('/music'))

    fireEvent.click(screen.getAllByRole('button', { name: 'Browse...' })[0])

    await waitFor(() => expect(screen.getByText(/no folder chooser is available here/)).toBeTruthy())
    fireEvent.change(input('Download folder'), { target: { value: '/typed' } })
    expect(input('Download folder').value).toBe('/typed')
  })

  it('shows the log folder only while a log is kept', async () => {
    serve({ 'GET /api/settings': { status: 200, body: document_ } })
    render(<SettingsForm runStarted={false} />)
    await waitFor(() => expect(input('Log folder').value).toBe('/logs'))

    fireEvent.click(screen.getByLabelText('Keep a log file'))

    expect(screen.queryByLabelText('Log folder')).toBeNull()
  })
})
