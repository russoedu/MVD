import type { SocketLike } from '@treactui/tty'
import { createWailsSocket } from '@treactui/tty'

const wails = createWailsSocket()

/** The connection the terminal opened, once it has. */
let current: SocketLike | undefined

/**
 * Created once: a new function on every render would make the terminal reconnect. It is
 * Wails' socket, kept so the page can send keys of its own (see sendTerminalInput).
 */
export const terminalSocket = (url?: string): SocketLike => {
  current = wails(url)

  return current
}

/** Sends bytes to the program as if they were typed in the terminal. */
export function sendTerminalInput (data: string): void {
  current?.send(JSON.stringify({ type: 'input', data }))
}
