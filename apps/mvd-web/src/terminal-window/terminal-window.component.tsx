import { TTY as Terminal } from '@treactui/tty'
import { terminalUrl } from './terminal-url.algorithm'

/**
 * The terminal-style interface: the whole window is mvd's own screens, the same
 * ones the terminal app shows, served over a WebSocket by the app.
 */
export function TerminalWindow () {
  return (
    <main className='terminal-window'>
      <Terminal url={terminalUrl(location)} label='MVD' />
    </main>
  )
}
