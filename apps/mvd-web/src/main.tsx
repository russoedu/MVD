import { lazy, StrictMode, Suspense } from 'react'
import * as ReactDOM from 'react-dom/client'
import { TerminalWindow } from './terminal-window'

// Loaded only on its own page, so the terminal page does not carry the classic themes.
const ThemePreview = lazy(async () => {
  const { ThemePreview: page } = await import('./theme-preview')

  return { default: page }
})

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement,
)

// The page is the terminal-style interface; the app serves it for every path. /themes shows the
// classic themes with a sample list, until the program can draw its own window in them (#345).
const page = location.pathname.replace(/\/+$/, '') === '/themes' ? <Suspense><ThemePreview /></Suspense> : <TerminalWindow />

root.render(
  <StrictMode>
    {page}
  </StrictMode>,
)
