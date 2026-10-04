import { requestJson } from '../api'
import type { PickedFolder, Settings, SettingsDocument } from './settings.contract'

export function loadSettings (): Promise<SettingsDocument> {
  return requestJson('GET', '/api/settings')
}

/** Saves, and returns what is now stored. Rejects with an ApiError whose `fields` name each bad input. */
export function saveSettings (settings: Settings): Promise<SettingsDocument> {
  return requestJson('PUT', '/api/settings', settings)
}

/** Opens the machine's folder chooser, starting at `start`. */
export function pickFolder (start: string): Promise<PickedFolder> {
  return requestJson('POST', '/api/folders/pick', { start })
}
