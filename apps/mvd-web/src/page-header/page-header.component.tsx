import { PixelLogo } from './pixel-logo.component'
import './page-header.style.css'

interface PageHeaderProps {
  onAbout: () => void
}

/** The banner above the terminal: pixel logo, scrolling tagline and the About button. */
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
    </header>
  )
}
