import type { AddSourcesResult } from './run-snapshot.contract'

interface ErrorBody { error?: string }

async function send<T> (path: string, body: unknown): Promise<T> {
  const response = await fetch(path, {
    method:  'POST',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify(body),
  })
  let payload = {} as T & ErrorBody
  try {
    payload = await response.json() as T & ErrorBody
  } catch {
    // An error from a proxy or a crash has no JSON body; the status says enough.
  }
  if (!response.ok) {
    throw new Error(payload.error ?? `The server answered ${response.status}`)
  }

  return payload
}

/** Queues what was pasted: one URL per line. */
export function addSources (text: string): Promise<AddSourcesResult> {
  return send('/api/sources', { text })
}

export function retryEntry (id: number): Promise<{ ok: boolean }> {
  return send(`/api/entries/${id}/retry`, {})
}

export function retryPlaylist (index: number): Promise<{ retried: number }> {
  return send(`/api/playlists/${index}/retry`, {})
}
