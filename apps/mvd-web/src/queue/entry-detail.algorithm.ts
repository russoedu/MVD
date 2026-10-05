import type { EntryView } from '../run'
import { formatBytes, formatEta, formatSpeed } from './format-progress.algorithm'

const ACTIVE = new Set(['resolving', 'downloading', 'merging'])

/** Whether the entry is being worked on right now, which is when it shows a progress bar. */
export function isActive (entry: EntryView): boolean {
  return ACTIVE.has(entry.state)
}

/** The second line of an entry: speed, time left and size while it runs, why it failed after. */
export function describeEntryDetail (entry: EntryView): string {
  if (!isActive(entry)) return entry.err

  return [formatSpeed(entry.speed), formatEta(entry.eta) && `ETA ${formatEta(entry.eta)}`, formatBytes(entry.totalBytes)]
    .filter(Boolean).join(' · ')
}
