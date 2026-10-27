import { useCallback, useEffect, useState } from 'react'
import type { PreferencesIntent, PreferencesSnapshot } from '../preferences-window'
import { onProgramEvent, sendToProgram } from './program-events.client'

/** The events of the preferences window (see apps/mvd/appwindow/classic_window_client.go). */
const prefsEvent = 'mvd:prefs'
const prefsIntentEvent = 'mvd:prefs-intent'

/**
 * The preferences while the program has them open, null while it has not, and the way to tell it
 * what the person changed.
 */
export function useLivePreferences (): [PreferencesSnapshot | null, (intent: PreferencesIntent) => void] {
  const [prefs, setPrefs] = useState<PreferencesSnapshot | null>(null)

  useEffect(() => onProgramEvent(prefsEvent, value => {
    if (typeof value !== 'string') return
    try {
      setPrefs(JSON.parse(value) as PreferencesSnapshot | null)
    } catch {
      // What cannot be read is left out; the next one replaces it.
    }
  }), [])

  const send = useCallback((intent: PreferencesIntent) => sendToProgram(prefsIntentEvent, JSON.stringify(intent)), [])

  return [prefs, send]
}
