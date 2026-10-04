import { useEffect, useState } from 'react'
import type { RunSnapshot } from './run-snapshot.contract'

export type Connection = 'connecting' | 'live' | 'lost'

export interface WatchedRun {
  snapshot:   RunSnapshot | undefined
  connection: Connection
}

/**
 * Follows the run over server-sent events. Each message is the whole snapshot, so
 * there is nothing to merge: a reconnect (the browser's own) just draws the next one.
 */
export function useRun (): WatchedRun {
  const [snapshot, setSnapshot] = useState<RunSnapshot>()
  const [connection, setConnection] = useState<Connection>('connecting')

  useEffect(() => {
    const source = new EventSource('/api/events')
    source.addEventListener('snapshot', event => {
      setSnapshot(JSON.parse((event as MessageEvent<string>).data) as RunSnapshot)
      setConnection('live')
    })
    source.addEventListener('error', () => {
      setConnection(source.readyState === EventSource.CLOSED ? 'lost' : 'connecting')
    })

    return () => source.close()
  }, [])

  return { snapshot, connection }
}
