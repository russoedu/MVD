/**
 * Whether an address from the terminal may be handed to the person's own browser:
 * only web addresses, so a link can never start another program.
 */
export function isWebAddress (address: string): boolean {
  try {
    const { protocol } = new URL(address)

    return protocol === 'http:' || protocol === 'https:'
  } catch {
    return false
  }
}

/** The question xterm asks before it follows a link; the app answers it itself. */
export function isLinkQuestion (message: string | undefined): boolean {
  return (message ?? '').startsWith('Do you want to navigate to ')
}
