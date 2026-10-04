import { requestJson } from '../api'
import type { AddSourcesResult } from './run-snapshot.contract'

/** Queues what was pasted: one URL per line. */
export function addSources (text: string): Promise<AddSourcesResult> {
  return requestJson('POST', '/api/sources', { text })
}

export function retryEntry (id: number): Promise<{ ok: boolean }> {
  return requestJson('POST', `/api/entries/${id}/retry`, {})
}

export function retryPlaylist (index: number): Promise<{ retried: number }> {
  return requestJson('POST', `/api/playlists/${index}/retry`, {})
}
