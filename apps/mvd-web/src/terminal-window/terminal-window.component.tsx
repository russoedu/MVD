import { TTY as Terminal } from '@treactui/tty'
import { useState } from 'react'
import { AboutThisApp } from '../about-this-app'
import { PageHeader } from '../page-header'
import { terminalUrl } from './terminal-url.algorithm'

/**
 * The page: a 90s style header above the terminal-style interface, which is
 * mvd's own screens, the same ones the terminal app shows, served over a
 * WebSocket by the app.
 */
export function TerminalWindow () {
  const [aboutOpen, setAboutOpen] = useState(false)

  return (
    <>
      <PageHeader onAbout={() => setAboutOpen(true)} />
      <main className='terminal-window'>
        <Terminal url={terminalUrl(location)} label='MVD' />
      </main>
      <AboutThisApp open={aboutOpen} onClose={() => setAboutOpen(false)} />
    </>
  )
}
