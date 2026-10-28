import { useEffect, useRef, useState } from 'react'

/**
 * The text of a field whose value lives in the program: every edit is sent there and comes back
 * in the next snapshot. The field shows what the person wrote at once; an answer to an edit the
 * person has already written past is skipped, or it would put the text back and drop the keys
 * typed since. What the program changes by itself (an emptied field) is taken as it comes.
 */
export function useEchoedText (text: string, send: (text: string) => void): [string, (next: string) => void] {
  const [draft, setDraft] = useState(text)
  const sent = useRef<string[]>([])

  useEffect(() => {
    const at = sent.current.indexOf(text)
    if (at === -1) {
      sent.current = []
      setDraft(text)
    } else {
      sent.current = sent.current.slice(at + 1)
    }
  }, [text])

  const write = (next: string): void => {
    sent.current.push(next)
    setDraft(next)
    send(next)
  }

  return [draft, write]
}
