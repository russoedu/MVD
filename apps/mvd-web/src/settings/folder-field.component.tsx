import { useState } from 'react'
import { pickFolder } from './settings-api.client'

interface FolderFieldProps {
  id:       string
  label:    string
  value:    string
  error?:   string
  onChange: (value: string) => void
}

/** A folder path that can be typed or chosen with the machine's own chooser. */
export function FolderField ({ id, label, value, error, onChange }: FolderFieldProps) {
  const [busy, setBusy] = useState(false)
  const [problem, setProblem] = useState('')

  async function browse () {
    setBusy(true)
    setProblem('')
    try {
      const picked = await pickFolder(value)
      if (!picked.cancelled) onChange(picked.path)
    } catch (error_) {
      setProblem(error_ instanceof Error ? error_.message : String(error_))
    } finally {
      setBusy(false)
    }
  }

  const message = error ?? problem

  return (
    <div className='field'>
      <label htmlFor={id}>{label}</label>
      <div className='folder-input'>
        <input id={id} type='text' value={value} onChange={event => onChange(event.target.value)} />
        <button type='button' disabled={busy} onClick={() => void browse()}>
          {busy ? 'Choosing...' : 'Browse...'}
        </button>
      </div>
      {message !== '' && <span className='message error'>{message}</span>}
    </div>
  )
}
