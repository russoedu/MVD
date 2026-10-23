import { PixelLogo } from './pixel-logo.component'
import { WindowControls } from './window-controls.component'
import './page-header.style.css'

interface PageHeaderProps {
  onAbout: () => void
}

/**
 * The banner above the terminal: pixel logo, scrolling tagline, the About button and the
 * window's own buttons. The window has no frame of the system's, so this is its title bar:
 * it is dragged to move the window (see page-header.style.css).
 */
export function PageHeader ({ onAbout }: PageHeaderProps) {
  return (
    <header className='page-header'>
      <PixelLogo />
      <div className='page-header-title'>
        <strong>Music Video Downloader</strong>
        <div className='page-header-ticker'>
          <span>*** I want my MTV... and I want it stored locally, total overload style! ***</span>
        </div>
      </div>
      <button type='button' className='page-header-button' onClick={onAbout}>
        About this app
      </button>
      <WindowControls />
    </header>
  )
}
