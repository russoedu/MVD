import { requestJson } from '../api'

/**
 * Asks MVD to remove itself. The app always asks the person again on their own screen, and
 * rejects with an ApiError (409) if they say no, so nothing is removed unless they agree there.
 */
export function uninstallApp (preferencesToo: boolean): Promise<{ status: string }> {
  return requestJson('POST', '/api/uninstall', { deletePreferences: preferencesToo })
}
