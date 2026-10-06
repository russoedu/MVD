import { StrictMode } from 'react'
import * as ReactDOM from 'react-dom/client'
import { TerminalWindow } from './terminal-window'

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement,
)

// The page is the terminal-style interface; the app serves it for every path.
root.render(
  <StrictMode>
    <TerminalWindow />
  </StrictMode>,
)
