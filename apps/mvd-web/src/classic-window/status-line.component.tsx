import { shortcutLabel, type ChromeLayout } from '../classic-themes'
import type { StatusSnapshot } from '../workspace-snapshot'
import { SizeGrip } from './size-grip.component'

interface StatusLineProps {
  status: StatusSnapshot
  layout: ChromeLayout
}

/**
 * The line at the foot of the window: how the run is going, the keys that do the most, and the
 * corner that resizes the window.
 */
export function StatusLine ({ status, layout }: StatusLineProps) {
  return (
    <div className='classic-status-line'>
      <span className='classic-status-summary' role='status'>{status.summary}</span>
      <span className='classic-status-hints'>
        {status.hints.map(hint => (
          <span key={hint.label} className='classic-status-hint'>
            <kbd>{shortcutLabel(hint.shortcut, layout.shortcutStyle)}</kbd> {hint.label}
          </span>
        ))}
      </span>
      <SizeGrip />
    </div>
  )
}
