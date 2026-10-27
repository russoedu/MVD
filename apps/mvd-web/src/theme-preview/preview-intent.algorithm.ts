import type { WorkspaceIntent, WorkspaceSnapshot } from '../workspace-snapshot'
import { sampleAdding, sampleSearch } from './sample-workspace.content'

/**
 * What the preview does with an intent, so the sample window can be tried: selecting, ticking,
 * and opening, filling and closing the window for adding music. It stands in for the program
 * only on the preview page; the real window will send intents to the program instead.
 */
export function applyPreviewIntent (snapshot: WorkspaceSnapshot, intent: WorkspaceIntent): WorkspaceSnapshot {
  switch (intent.kind) {
    case 'select-source': {
      return { ...snapshot, sources: { ...snapshot.sources, selected: intent.index } }
    }
    case 'select-song': {
      return { ...snapshot, songs: { ...snapshot.songs, selected: intent.index } }
    }
    case 'tick-song': {
      return {
        ...snapshot,
        songs: {
          ...snapshot.songs,
          rows: snapshot.songs.rows.map((row, index) => index === intent.index ? { ...row, ticked: intent.ticked } : row),
        },
      }
    }
    case 'menu':
    case 'button': {
      return intent.id === 'add-source' ? { ...snapshot, adding: sampleAdding } : snapshot
    }
    case 'adding-mode': {
      if (!snapshot.adding) return snapshot
      const fields = intent.mode === 'url' ? { fields: sampleAdding.fields, help: sampleAdding.help } : sampleSearch

      return { ...snapshot, adding: { ...snapshot.adding, mode: intent.mode, ...fields, problem: undefined } }
    }
    case 'adding-text': {
      if (!snapshot.adding) return snapshot

      return {
        ...snapshot,
        adding: {
          ...snapshot.adding,
          fields: snapshot.adding.fields.map(field => field.field === intent.field ? { ...field, text: intent.text } : field),
        },
      }
    }
    case 'adding-button': {
      return addingButton(snapshot, intent.id)
    }
    case 'window':
    case 'key': {
      return snapshot
    }
  }
}

function addingButton (snapshot: WorkspaceSnapshot, id: string): WorkspaceSnapshot {
  const adding = snapshot.adding
  if (!adding || id === 'cancel') return { ...snapshot, adding: undefined }
  if (adding.fields.every(field => field.text.trim() === '')) {
    return { ...snapshot, adding: { ...adding, problem: adding.mode === 'url' ? 'Paste a link' : 'Write the artist' } }
  }

  return { ...snapshot, adding: undefined }
}
