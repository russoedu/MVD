import type { SocketLike } from '@treactui/tty'

/**
 * The frame that tells the terminal to hide its cursor (DECTCEM). The program never sends it, and
 * a screen made of boxes and key hints has no use for a blinking block at the last cell it drew;
 * the text fields draw their own cursor.
 */
const hideCursor = JSON.stringify({ type: 'output', data: '\u{1B}[?25l' })

/**
 * Wraps a connection so the terminal hides its cursor before the first frame of the program reaches
 * it. Everything else passes through as it is.
 */
export function withHiddenCursor (socket: SocketLike): SocketLike {
  return {
    // Called on the socket, not copied: a real WebSocket needs its own `this`.
    send:  frame => { socket.send(frame) },
    close: () => { socket.close() },
    addEventListener (type, listener) {
      if (type !== 'message') {
        socket.addEventListener(type, listener)

        return
      }

      let first = true
      socket.addEventListener('message', event => {
        if (first) {
          first = false
          listener({ data: hideCursor })
        }
        listener(event)
      })
    },
  }
}
