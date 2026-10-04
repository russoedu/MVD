import type { AddSourcesResult } from '../run'

/** Describes what became of the lines that were sent, or '' if nothing to say. */
export function describeResult (result: AddSourcesResult): string {
  const parts: string[] = []
  if (result.added.length > 0) parts.push(`${result.added.length} added`)
  if (result.duplicates.length > 0) parts.push(`${result.duplicates.length} already queued`)
  if (result.rejected.length > 0) parts.push(`${result.rejected.length} not a web link`)

  return parts.join(', ')
}
