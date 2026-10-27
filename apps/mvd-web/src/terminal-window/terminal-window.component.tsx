import { TTY as Terminal } from '@treactui/tty'
import { useRef, useState } from 'react'
import { AboutThisApp } from '../about-this-app'
import { PageHeader } from '../page-header'
import { useCaptureBrowserShortcuts } from './capture-browser-shortcuts.hook'
import { useDimWhileModal } from './dim-while-modal.hook'
import { useFocusTerminal } from './focus-terminal.hook'
import { useFollowTerminalBackground } from './follow-terminal-background.hook'
import { useOpenLinksInBrowser } from './open-links-in-browser.hook'
import { useForwardShiftClicks } from './shift-click.hook'
import { terminalSocket } from './terminal-socket.client'

/**
 * How the terminal is shown: as the page, with its 90s header; behind a classic window, out of
 * sight and out of reach of the keyboard, still running MVD; or inset in a classic window's
 * dialog, for a screen the classic themes do not draw yet (the dialog says where, see
 * --terminal-inset-* in the style).
 */
export type TerminalLook = 'page' | 'behind' | 'inset'

interface TerminalWindowProps {
  look?: TerminalLook
}

/**
 * The page: a 90s style header above the terminal-style interface, which is
 * mvd's own screens, the same ones the terminal app shows. It talks to Go through
 * Wails' events, so it only works inside the app's window. It stays mounted whatever the
 * window style, so the program it is connected to keeps its one connection.
 */
export function TerminalWindow ({ look = 'page' }: TerminalWindowProps) {
  const [aboutOpen, setAboutOpen] = useState(false)
  const terminalRef = useRef<HTMLElement>(null)
  const page = look === 'page'
  const reachable = look !== 'behind'
  useCaptureBrowserShortcuts(reachable)
  useDimWhileModal(page)
  useOpenLinksInBrowser()
  useFollowTerminalBackground(terminalRef)
  useFocusTerminal(terminalRef, reachable)
  useForwardShiftClicks(terminalRef)

  return (
    <>
      {page && <PageHeader onAbout={() => setAboutOpen(true)} />}
      <main ref={terminalRef} className='terminal-window' data-look={look} inert={!reachable}>
        <Terminal createSocket={terminalSocket} label='MVD' />
      </main>
      <AboutThisApp open={aboutOpen} onClose={() => setAboutOpen(false)} />
    </>
  )
}
