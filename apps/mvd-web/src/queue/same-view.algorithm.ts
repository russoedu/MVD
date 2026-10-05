import type { EntryView, PlaylistView } from '../run'

function samePrimitives<T extends object> (left: T, right: T, skip?: keyof T): boolean {
  if (left === right) return true

  return (Object.keys(left) as Array<keyof T>).every(key => key === skip || left[key] === right[key])
}

/** Every snapshot is new objects, so rows are compared by value, field by field. */
export function sameEntry (left: EntryView, right: EntryView): boolean {
  return samePrimitives(left, right)
}

/** A playlist heading does not show its entry ids, so only the other fields count. */
export function samePlaylist (left: PlaylistView, right: PlaylistView): boolean {
  return samePrimitives(left, right, 'entries')
}
