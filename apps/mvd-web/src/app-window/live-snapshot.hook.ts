import { useCallback, useEffect, useState } from 'react'
import type { WorkspaceIntent, WorkspaceSnapshot } from '../workspace-snapshot'
import { onProgramEvent, sendToProgram } from './program-events.client'

/** The events of the classic window (see apps/mvd/appwindow/classic_window_client.go). */
const snapshotEvent = 'mvd:snapshot'
const intentEvent = 'mvd:intent'

/**
 * What the program says its window shows, and the way to tell it what the person did. The
 * snapshot is null until the program has sent one; each one replaces the last.
 */
export function useLiveSnapshot (): [WorkspaceSnapshot | null, (intent: WorkspaceIntent) => void] {
  const [snapshot, setSnapshot] = useState<WorkspaceSnapshot | null>(null)

  useEffect(() => onProgramEvent(snapshotEvent, value => {
    if (typeof value !== 'string') return
    try {
      setSnapshot(JSON.parse(value) as WorkspaceSnapshot)
    } catch {
      // A snapshot that cannot be read is left out; the next one replaces it.
    }
  }), [])

  const send = useCallback((intent: WorkspaceIntent) => sendToProgram(intentEvent, JSON.stringify(intent)), [])

  return [snapshot, send]
}
