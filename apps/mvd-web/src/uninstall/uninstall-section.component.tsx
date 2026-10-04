import { useState } from 'react'
import { uninstallApp } from './uninstall-api.client'

type Step = 'idle' | 'confirming' | 'asking' | 'removed'

export function UninstallSection () {
  const [step, setStep] = useState<Step>('idle')
  const [preferencesToo, setPreferencesToo] = useState(false)
  const [error, setError] = useState('')

  async function remove () {
    setStep('asking')
    setError('')
    try {
      await uninstallApp(preferencesToo)
      setStep('removed')
    } catch (error_) {
      setError(error_ instanceof Error ? error_.message : String(error_))
      setStep('confirming')
    }
  }

  if (step === 'removed') {
    return (
      <section className='remove-app'>
        <p role='status'>MVD has been removed. You can close this page.</p>
      </section>
    )
  }

  return (
    <section className='remove-app'>
      <h2>Remove MVD</h2>
      {step === 'idle' && (
        <>
          <p className='message'>Uninstall MVD from this computer. Your downloaded videos are never deleted.</p>
          <button type='button' className='danger' onClick={() => setStep('confirming')}>Remove MVD...</button>
        </>
      )}
      {step !== 'idle' && (
        <>
          <p className='message'>
            MVD will ask you once more in a window on this computer, showing exactly what it will remove.
            Your downloaded videos, and the folder they are in, are never deleted.
          </p>
          <label className='check'>
            <input type='checkbox' checked={preferencesToo} disabled={step === 'asking'} onChange={event => setPreferencesToo(event.target.checked)} />
            Also delete my preferences (settings, list and the downloaded tools)
          </label>
          <div className='settings-actions'>
            <button type='button' className='danger' disabled={step === 'asking'} onClick={() => void remove()}>
              {step === 'asking' ? 'Waiting for your answer...' : 'Remove MVD'}
            </button>
            <button type='button' className='secondary' disabled={step === 'asking'} onClick={() => setStep('idle')}>Cancel</button>
          </div>
          {step === 'asking' && <p className='message' role='status'>Answer the question in the window that opened on this computer.</p>}
          {error !== '' && <p className='message error' role='alert'>{error}</p>}
        </>
      )}
    </section>
  )
}
