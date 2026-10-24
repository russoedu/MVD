import { TTY as Terminal } from '@treactui/tty'
import { useRef, useState } from 'react'
import { AboutThisApp } from '../about-this-app'
import { PageHeader } from '../page-header'
import { useCaptureBrowserShortcuts } from './capture-browser-shortcuts.hook'
import { useDimWhileModal } from './dim-while-modal.hook'
import { useFocusTerminal } from './focus-terminal.hook'
import { useFollowTerminalBackground } from './follow-terminal-background.hook'
import { useOpenLinksInBrowser } from './open-links-in-browser.hook'
import { terminalSocket } from './terminal-socket.client'

/**
 * The page: a 90s style header above the terminal-style interface, which is
 * mvd's own screens, the same ones the terminal app shows. It talks to Go through
 * Wails' events, so it only works inside the app's window.
 */
export function TerminalWindow () {
  const [aboutOpen, setAboutOpen] = useState(false)
  const terminalRef = useRef<HTMLElement>(null)
  useCaptureBrowserShortcuts()
  useDimWhileModal()
  useOpenLinksInBrowser()
  useFollowTerminalBackground(terminalRef)
  useFocusTerminal(terminalRef)

  return (
    <>
      <PageHeader onAbout={() => setAboutOpen(true)} />
      <main ref={terminalRef} className='terminal-window'>
        <Terminal createSocket={terminalSocket} label='MVD' />
      </main>
      <AboutThisApp open={aboutOpen} onClose={() => setAboutOpen(false)} />
    </>
  )
}
