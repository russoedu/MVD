import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { UninstallSection } from './uninstall-section.component'

const fetchMock = jest.fn()

type Reply = { status: number, body: unknown }

function serve (reply: Reply | (() => Promise<Reply>)) {
  fetchMock.mockImplementation(async () => {
    const { status, body } = typeof reply === 'function' ? await reply() : reply

    return { ok: status < 400, status, json: () => Promise.resolve(body) }
  })
}

const bodyOf = (call: unknown[]) => JSON.parse((call[1] as RequestInit).body as string) as Record<string, unknown>

beforeEach(() => {
  fetchMock.mockReset()
  Object.defineProperty(globalThis, 'fetch', { value: fetchMock, configurable: true, writable: true })
})

describe('UninstallSection', () => {
  it('only offers the removal at first and calls nothing', () => {
    render(<UninstallSection />)

    expect(screen.getByRole('button', { name: 'Remove MVD...' })).toBeTruthy()
    expect(screen.getByText(/downloaded videos are never deleted/)).toBeTruthy()
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('keeps the preferences unless the box is ticked', async () => {
    serve({ status: 202, body: { status: 'removing' } })
    render(<UninstallSection />)

    fireEvent.click(screen.getByRole('button', { name: 'Remove MVD...' }))
    expect(screen.getByLabelText<HTMLInputElement>(/Also delete my preferences/).checked).toBe(false)
    fireEvent.click(screen.getByRole('button', { name: 'Remove MVD' }))

    await waitFor(() => expect(screen.getByRole('status').textContent).toContain('MVD has been removed'))
    expect(fetchMock.mock.calls[0][0]).toBe('/api/uninstall')
    expect((fetchMock.mock.calls[0][1] as RequestInit).method).toBe('POST')
    expect(bodyOf(fetchMock.mock.calls[0])).toEqual({ deletePreferences: false })
  })

  it('asks for the preferences to go too when the box is ticked', async () => {
    serve({ status: 202, body: { status: 'removing' } })
    render(<UninstallSection />)

    fireEvent.click(screen.getByRole('button', { name: 'Remove MVD...' }))
    fireEvent.click(screen.getByLabelText(/Also delete my preferences/))
    fireEvent.click(screen.getByRole('button', { name: 'Remove MVD' }))

    await waitFor(() => expect(bodyOf(fetchMock.mock.calls[0])).toEqual({ deletePreferences: true }))
    await waitFor(() => expect(screen.getByRole('status').textContent).toContain('You can close this page'))
  })

  it('says it is waiting for the answer on the computer while the question is open', async () => {
    const answer = Promise.withResolvers<Reply>()
    serve(() => answer.promise)
    render(<UninstallSection />)

    fireEvent.click(screen.getByRole('button', { name: 'Remove MVD...' }))
    fireEvent.click(screen.getByRole('button', { name: 'Remove MVD' }))

    await waitFor(() => expect(screen.getByRole('status').textContent).toContain('window that opened on this computer'))
    expect(screen.getByRole('button', { name: 'Waiting for your answer...' }).hasAttribute('disabled')).toBe(true)
    answer.resolve({ status: 202, body: { status: 'removing' } })
    await waitFor(() => expect(screen.getByRole('status').textContent).toContain('has been removed'))
  })

  it('shows the reason and goes back to the choice when the person says no on their screen', async () => {
    serve({ status: 409, body: { error: 'the removal was cancelled; nothing was removed' } })
    render(<UninstallSection />)

    fireEvent.click(screen.getByRole('button', { name: 'Remove MVD...' }))
    fireEvent.click(screen.getByRole('button', { name: 'Remove MVD' }))

    await waitFor(() => expect(screen.getByRole('alert').textContent).toContain('nothing was removed'))
    expect(screen.getByRole('button', { name: 'Remove MVD' }).hasAttribute('disabled')).toBe(false)
    expect(screen.queryByText(/has been removed/)).toBeNull()
  })

  it('can be cancelled before anything is sent', () => {
    render(<UninstallSection />)

    fireEvent.click(screen.getByRole('button', { name: 'Remove MVD...' }))
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))

    expect(screen.getByRole('button', { name: 'Remove MVD...' })).toBeTruthy()
    expect(fetchMock).not.toHaveBeenCalled()
  })
})
