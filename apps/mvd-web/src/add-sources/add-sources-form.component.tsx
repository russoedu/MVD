import { useState } from 'react'
import type { FormEvent } from 'react'
import { addSources } from '../run'
import { describeResult } from './describe-add-result.algorithm'

export function AddSourcesForm () {
  const [text, setText] = useState('')
  const [busy, setBusy] = useState(false)
  const [message, setMessage] = useState('')
  const [failed, setFailed] = useState(false)

  async function submit (event: FormEvent) {
    event.preventDefault()
    if (text.trim() === '') return
    setBusy(true)
    try {
      const result = await addSources(text)
      setMessage(describeResult(result))
      setFailed(false)
      // Keep what was refused so it can be fixed, drop what was taken.
      setText(result.rejected.join('\n'))
    } catch (error) {
      setMessage(error instanceof Error ? error.message : String(error))
      setFailed(true)
    } finally {
      setBusy(false)
    }
  }

  return (
    <form className='add-sources' onSubmit={event => void submit(event)}>
      <label htmlFor='sources'>Paste playlist or video links, one per line</label>
      <textarea
        id='sources'
        rows={4}
        value={text}
        placeholder='https://www.youtube.com/playlist?list=...'
        onChange={event => setText(event.target.value)}
      />
      <div className='add-sources-actions'>
        <button type='submit' disabled={busy || text.trim() === ''}>
          {busy ? 'Adding...' : 'Add to queue'}
        </button>
        <span role='status' className={failed ? 'message error' : 'message'}>{message}</span>
      </div>
    </form>
  )
}
