import { askWindow } from '../window-actions'
import './window-controls.style.css'

/**
 * The buttons of the window, in the 90s style, for a window with no frame of the system's:
 * the header is its title bar, and these are minimise, maximise and close.
 */
export function WindowControls () {
  return (
    <div className='window-controls'>
      <button type='button' className='window-control' aria-label='Minimise' title='Minimise' onClick={() => askWindow('minimise')}>_</button>
      <button type='button' className='window-control' aria-label='Maximise' title='Maximise' onClick={() => askWindow('maximise')}>□</button>
      <button type='button' className='window-control window-control-close' aria-label='Close' title='Close' onClick={() => askWindow('close')}>×</button>
    </div>
  )
}
