import { ProgressBar as AriaProgressBar } from 'react-aria-components'
import type { Progress } from '../workspace-snapshot'

interface ProgressBarProps {
  progress: Progress
  label:    string
}

/** A song's progress in the theme's own bar; "unknown" draws the theme's moving bar. Nothing when there is none. */
export function ProgressBar ({ progress, label }: ProgressBarProps) {
  if (progress === null) return null
  const unknown = progress === 'unknown'

  return (
    <AriaProgressBar
      className='classic-progress'
      aria-label={label}
      value={unknown ? undefined : progress * 100}
      isIndeterminate={unknown}
    >
      {({ percentage }) => (
        <div className='classic-progress-trough'>
          <div
            className='classic-progress-fill'
            data-unknown={unknown || undefined}
            style={{ width: unknown ? '100%' : `${percentage ?? 0}%` }}
          />
        </div>
      )}
    </AriaProgressBar>
  )
}
