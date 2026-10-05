/** The part of `window.location` the endpoint address is built from. */
export interface PageLocation {
  protocol: string
  host:     string
}

/** Where the terminal interface's WebSocket is, on the site that served the page. */
export function terminalUrl (location: PageLocation): string {
  const scheme = location.protocol === 'https:' ? 'wss' : 'ws'

  return `${scheme}://${location.host}/term`
}
