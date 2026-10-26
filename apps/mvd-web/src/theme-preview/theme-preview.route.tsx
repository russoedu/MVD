import { useState, type ReactNode } from 'react'
import { AddMusicWindow } from '../add-music-window'
import { classicThemeOf, classicThemes } from '../classic-themes'
import { DetailsPane, SongsTable, SourcesPane } from '../classic-panes'
import { ClassicWindow } from '../classic-window'
import { askWindow } from '../window-actions'
import type { WorkspaceIntent } from '../workspace-snapshot'
import { applyPreviewIntent } from './preview-intent.algorithm'
import { sampleWorkspace } from './sample-workspace.content'
import './theme-preview.style.css'

/** The theme named in the address (?theme=aqua), or the first one. */
function themeInAddress () {
  return classicThemeOf(new URLSearchParams(globalThis.location?.search ?? '').get('theme'))
}

/**
 * The page at /themes: MVD's main window drawn with a sample list in each classic theme, to see
 * and try the themes before the program sends its own snapshot (#345).
 */
export function ThemePreview () {
  const [theme, setTheme] = useState(themeInAddress)
  const [snapshot, setSnapshot] = useState(sampleWorkspace)
  const [last, setLast] = useState('')

  const onIntent = (intent: WorkspaceIntent) => {
    setLast(JSON.stringify(intent))
    if (intent.kind === 'window') askWindow(intent.action)
    setSnapshot(current => applyPreviewIntent(current, intent))
  }

  return (
    <div className='theme-preview'>
      <div className='theme-preview-bar' role='group' aria-label='Theme'>
        {classicThemes.map(each => (
          <button
            key={each.id}
            type='button'
            aria-pressed={each.id === theme.id}
            onClick={() => {
              setTheme(each)
              globalThis.history?.replaceState(null, '', `?theme=${each.id}`)
            }}
          >
            {each.label}
          </button>
        ))}
        <output className='theme-preview-intent'>{last}</output>
      </div>
      <div className='theme-preview-desktop'>
        <ClassicWindow
          theme={theme}
          snapshot={snapshot}
          onIntent={onIntent}
          overlay={(container, layout): ReactNode => snapshot.adding && (
            <AddMusicWindow
              adding={snapshot.adding}
              layout={layout}
              container={container}
              onMode={mode => onIntent({ kind: 'adding-mode', mode })}
              onText={(field, text) => onIntent({ kind: 'adding-text', field, text })}
              onButton={id => onIntent({ kind: 'adding-button', id })}
            />
          )}
        >
          <SourcesPane sources={snapshot.sources} onSelect={index => onIntent({ kind: 'select-source', index })} />
          <SongsTable
            songs={snapshot.songs}
            onSelect={index => onIntent({ kind: 'select-song', index })}
            onTick={(index, ticked) => onIntent({ kind: 'tick-song', index, ticked })}
          />
          <DetailsPane details={snapshot.details} />
        </ClassicWindow>
      </div>
    </div>
  )
}
