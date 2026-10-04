import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { AddSourcesForm } from './add-sources-form.component'

const fetchMock = jest.fn()

beforeEach(() => {
  fetchMock.mockReset()
  Object.defineProperty(globalThis, 'fetch', { value: fetchMock, configurable: true, writable: true })
})

function respond (status: number, body: unknown) {
  fetchMock.mockResolvedValue({ ok: status < 400, status, json: () => Promise.resolve(body) })
}

describe('AddSourcesForm', () => {
  it('does nothing while the box is empty', () => {
    render(<AddSourcesForm />)
    expect(screen.getByRole<HTMLButtonElement>('button').disabled).toBe(true)
  })

  it('sends the pasted text, reports the outcome and clears what was taken', async () => {
    respond(200, { added: ['https://a.example/1'], duplicates: [], rejected: [] })
    render(<AddSourcesForm />)

    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'https://a.example/1' } })
    fireEvent.click(screen.getByRole('button'))

    await waitFor(() => expect(screen.getByRole('status').textContent).toBe('1 added'))
    expect(screen.getByRole<HTMLTextAreaElement>('textbox').value).toBe('')
    const [path, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(path).toBe('/api/sources')
    expect(init.method).toBe('POST')
    expect(new Headers(init.headers).get('Content-Type')).toBe('application/json')
    expect(JSON.parse(init.body as string)).toEqual({ text: 'https://a.example/1' })
  })

  it('keeps the lines that were refused so they can be fixed', async () => {
    respond(200, { added: [], duplicates: [], rejected: ['not a link'] })
    render(<AddSourcesForm />)

    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'not a link' } })
    fireEvent.click(screen.getByRole('button'))

    await waitFor(() => expect(screen.getByRole('status').textContent).toBe('1 not a web link'))
    expect(screen.getByRole<HTMLTextAreaElement>('textbox').value).toBe('not a link')
  })

  it('shows the server\'s own message when it fails, and keeps the text', async () => {
    respond(502, { error: 'yt-dlp is missing' })
    render(<AddSourcesForm />)

    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'https://a.example/1' } })
    fireEvent.click(screen.getByRole('button'))

    await waitFor(() => expect(screen.getByRole('status').textContent).toBe('yt-dlp is missing'))
    expect(screen.getByRole<HTMLTextAreaElement>('textbox').value).toBe('https://a.example/1')
  })
})
