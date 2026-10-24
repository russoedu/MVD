import type { SocketLike } from '@treactui/tty'
import { withHiddenCursor } from './hidden-cursor-socket.client'

type Listener = (event: { data?: unknown }) => void

function fakeSocket () {
  const listeners: Record<string, Listener[]> = { open: [], close: [], message: [] }
  const sent: string[] = []
  const socket: SocketLike = {
    send:  frame => { sent.push(frame) },
    close: () => {},
    addEventListener (type, listener) { listeners[type]?.push(listener) },
  }
  const emit = (type: string, data?: unknown) => {
    const heard = listeners[type] ?? []

    for (const listener of heard) {
      listener({ data })
    }
  }

  return { socket, emit, sent }
}

describe('withHiddenCursor', () => {
  it('hides the cursor before the first frame of the program', () => {
    const { socket, emit } = fakeSocket()
    const received: unknown[] = []
    withHiddenCursor(socket).addEventListener('message', event => { received.push(event.data) })

    emit('message', '{"type":"output","data":"hello"}')

    expect(received).toHaveLength(2)
    expect(JSON.parse(received[0] as string)).toEqual({ type: 'output', data: '\u{1B}[?25l' })
    expect(received[1]).toBe('{"type":"output","data":"hello"}')
  })

  it('does it once per connection, not before every frame', () => {
    const { socket, emit } = fakeSocket()
    const received: unknown[] = []
    withHiddenCursor(socket).addEventListener('message', event => { received.push(event.data) })

    emit('message', 'one')
    emit('message', 'two')

    expect(received).toHaveLength(3)
    expect(received.slice(1)).toEqual(['one', 'two'])
  })

  it('leaves open and close, and what is sent, as they are', () => {
    const { socket, emit, sent } = fakeSocket()
    const wrapped = withHiddenCursor(socket)
    const events: string[] = []
    wrapped.addEventListener('open', () => { events.push('open') })
    wrapped.addEventListener('close', () => { events.push('close') })

    emit('open')
    emit('close')
    wrapped.send('input')

    expect(events).toEqual(['open', 'close'])
    expect(sent).toEqual(['input'])
  })

  it('calls send and close on the socket itself, as a real WebSocket needs', () => {
    class NativeLike implements SocketLike {
      readonly sent: string[] = []
      closed = false
      send (frame: string): void { this.sent.push(frame) }
      close (): void { this.closed = true }
      addEventListener (): void { /* nothing to hear */ }
    }
    const native = new NativeLike()
    const wrapped = withHiddenCursor(native)

    wrapped.send('input')
    wrapped.close()

    expect(native.sent).toEqual(['input'])
    expect(native.closed).toBe(true)
  })
})
