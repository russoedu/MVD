import { splitMnemonic } from '../classic-themes'
import type { Mnemonic } from '../workspace-snapshot'

interface MnemonicLabelProps {
  label:     string
  mnemonic?: Mnemonic
}

/**
 * A label with its keyboard letter marked, for the theme to underline: always on Windows 95,
 * once Alt is pressed on Vista, never on a Mac.
 */
export function MnemonicLabel ({ label, mnemonic }: MnemonicLabelProps) {
  const parts = splitMnemonic(label, mnemonic)
  if (!parts) return label

  // One span around it all, so a flex parent (a radio, a button) keeps it as one piece of text.
  return <span>{parts.before}<span className='classic-mnemonic'>{parts.letter}</span>{parts.after}</span>
}
