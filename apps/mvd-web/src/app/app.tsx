import { AddSourcesForm } from '../add-sources/add-sources-form.component'
import { QueueView } from '../queue/queue-view.component'
import { useRun } from '../run'

const CONNECTION_TEXT = {
  connecting: 'Connecting...',
  live:       '',
  lost:       'Lost contact with MVD. Is it still running?',
} as const

export function App () {
  const { snapshot, connection } = useRun()

  return (
    <main>
      <h1>MVD</h1>
      {CONNECTION_TEXT[connection] !== '' && <p className='message error' role='alert'>{CONNECTION_TEXT[connection]}</p>}
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
    </main>
  )
}

export default App
