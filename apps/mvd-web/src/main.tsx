import { StrictMode } from 'react'
import * as ReactDOM from 'react-dom/client'
import App from './app/app'
import { TerminalWindow } from './terminal-window'

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement,
)

// The app serves one page for every path; /terminal is the terminal-style interface.
root.render(
  <StrictMode>
    {location.pathname === '/terminal' ? <TerminalWindow /> : <App />}
  </StrictMode>,
)
