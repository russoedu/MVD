import { createWailsSocket, TTY as Terminal } from '@treactui/tty'
import { useState } from 'react'
import { AboutThisApp } from '../about-this-app'
import { PageHeader } from '../page-header'
import { useCaptureBrowserShortcuts } from './capture-browser-shortcuts.hook'

/** Created once: a new function on every render would make the terminal reconnect. */
const wailsSocket = createWailsSocket()

/**
 * The page: a 90s style header above the terminal-style interface, which is
 * mvd's own screens, the same ones the terminal app shows. It talks to Go through
 * Wails' events, so it only works inside the app's window.
 */
export function TerminalWindow () {
  const [aboutOpen, setAboutOpen] = useState(false)
  useCaptureBrowserShortcuts()

  return (
    <>
      <PageHeader onAbout={() => setAboutOpen(true)} />
      <main className='terminal-window'>
        <Terminal createSocket={wailsSocket} label='MVD' />
      </main>
      <AboutThisApp open={aboutOpen} onClose={() => setAboutOpen(false)} />
    </>
  )
}
