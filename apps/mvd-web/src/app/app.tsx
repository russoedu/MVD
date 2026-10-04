import { useState } from 'react'
import { AddSourcesForm } from '../add-sources'
import { QueueView } from '../queue'
import { useRun } from '../run'
import { SettingsForm } from '../settings'
import { UninstallSection } from '../uninstall'

const CONNECTION_TEXT = {
  connecting: 'Connecting...',
  live:       '',
  lost:       'Lost contact with MVD. Is it still running?',
} as const

export function App () {
  const { snapshot, connection } = useRun()
  const [view, setView] = useState<'queue' | 'settings'>('queue')

  return (
    <main>
      <header className='app-header'>
        <h1>MVD</h1>
        <nav>
          <button type='button' className='tab' aria-pressed={view === 'queue'} onClick={() => setView('queue')}>Queue</button>
          <button type='button' className='tab' aria-pressed={view === 'settings'} onClick={() => setView('settings')}>Settings</button>
        </nav>
      </header>
      {CONNECTION_TEXT[connection] !== '' && <p className='message error' role='alert'>{CONNECTION_TEXT[connection]}</p>}

      <section hidden={view !== 'queue'}>
        <AddSourcesForm />
        {snapshot && (
          <>
            <p className='tally'>
              {snapshot.tally.done} done · {snapshot.tally.running} running · {snapshot.tally.queued} queued
              {snapshot.tally.failed > 0 && ` · ${snapshot.tally.failed} failed`}
            </p>
            <QueueView snapshot={snapshot} />
          </>
        )}
      </section>

      {view === 'settings' && (
        <>
          <SettingsForm runStarted={(snapshot?.playlists.length ?? 0) > 0} />
          <UninstallSection />
        </>
      )}
    </main>
  )
}

export default App
