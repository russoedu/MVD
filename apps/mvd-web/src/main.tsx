import { lazy, StrictMode, Suspense } from 'react'
import * as ReactDOM from 'react-dom/client'
import { AppWindow } from './app-window'
import { installProgramRuntime } from './program-runtime'

// Loaded only on its own page, so the terminal page does not carry the classic themes.
const ThemePreview = lazy(async () => {
  const { ThemePreview: page } = await import('./theme-preview')

  return { default: page }
})

installProgramRuntime()

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement,
)

// The page is MVD's window, in the style the preferences say; the app serves it for every path.
// /themes shows the classic themes with a sample list, to try them without the program.
const page = location.pathname.replace(/\/+$/, '') === '/themes' ? <Suspense><ThemePreview /></Suspense> : <AppWindow />

root.render(
  <StrictMode>
    {page}
  </StrictMode>,
)
