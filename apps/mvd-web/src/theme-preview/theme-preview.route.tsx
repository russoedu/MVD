import { useState } from 'react'
import { ClassicApp } from '../classic-app'
import { classicThemeOf, classicThemes } from '../classic-themes'
import { askWindow } from '../window-actions'
import type { Activity, WorkspaceIntent } from '../workspace-snapshot'
import { applyPreviewIntent } from './preview-intent.algorithm'
import { sampleWorkspace } from './sample-workspace.content'
import './theme-preview.style.css'

/** What the pointer can be asked to show, to try each theme's busy pointers. */
const activities: { activity: Activity, label: string }[] = [
  { activity: 'idle', label: 'Idle' },
  { activity: 'background', label: 'Working behind' },
  { activity: 'waiting', label: 'Keeping you waiting' },
]

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
  const [scheme, setScheme] = useState('')
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
              setScheme('')
              globalThis.history?.replaceState(null, '', `?theme=${each.id}`)
            }}
          >
            {each.label}
          </button>
        ))}
        <label className='theme-preview-activity'>
          Scheme
          <select value={scheme} onChange={event => setScheme(event.target.value)}>
            {(theme.schemes ?? []).map((each, index) => <option key={each.id} value={index === 0 ? '' : each.id}>{each.label}</option>)}
          </select>
        </label>
        <label className='theme-preview-activity'>
          Pointer
          <select
            value={snapshot.activity}
            onChange={event => setSnapshot(current => ({ ...current, activity: event.target.value as Activity }))}
          >
            {activities.map(({ activity, label }) => <option key={activity} value={activity}>{label}</option>)}
          </select>
        </label>
        <output className='theme-preview-intent'>{last}</output>
      </div>
      <div className='theme-preview-desktop'>
        <ClassicApp theme={theme} scheme={scheme || undefined} snapshot={snapshot} onIntent={onIntent} />
      </div>
    </div>
  )
}
