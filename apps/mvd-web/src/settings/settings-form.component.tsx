import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { ApiError } from '../api'
import { FolderField } from './folder-field.component'
import { loadSettings, saveSettings } from './settings-api.client'
import type { Settings, SettingsOptions } from './settings.contract'

interface SettingsFormProps {
  /** True once downloads have started: the running engine keeps the settings it started with. */
  runStarted: boolean
}

function Select ({ id, label, value, choices, error, onChange }: {
  id:       string
  label:    string
  value:    string
  choices:  { value: string, label: string }[]
  error?:   string
  onChange: (value: string) => void
}) {
  return (
    <div className='field'>
      <label htmlFor={id}>{label}</label>
      <select id={id} value={value} onChange={event => onChange(event.target.value)}>
        {choices.map(choice => <option key={choice.value} value={choice.value}>{choice.label}</option>)}
      </select>
      {error && <span className='message error'>{error}</span>}
    </div>
  )
}

const same = (values: string[]) => values.map(value => ({ value, label: value }))

export function SettingsForm ({ runStarted }: SettingsFormProps) {
  const [settings, setSettings] = useState<Settings>()
  const [options, setOptions] = useState<SettingsOptions>()
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [message, setMessage] = useState('')
  const [failed, setFailed] = useState(false)
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    async function load () {
      try {
        const loaded = await loadSettings()
        setSettings(loaded.settings)
        setOptions(loaded.options)
      } catch (error) {
        setMessage(error instanceof Error ? error.message : String(error))
        setFailed(true)
      }
    }
    void load()
  }, [])

  if (!settings || !options) {
    return <p className={failed ? 'message error' : 'message'} role='status'>{message === '' ? 'Loading settings...' : message}</p>
  }

  const update = <K extends keyof Settings>(key: K, value: Settings[K]) => {
    setSettings({ ...settings, [key]: value })
  }
  const number = (key: 'maxConcurrentDownloads' | 'concurrentFragments') => (text: string) => {
    update(key, text === '' ? 0 : Number(text))
  }

  async function submit (event: FormEvent) {
    event.preventDefault()
    if (!settings) return
    setBusy(true)
    try {
      const saved = await saveSettings(settings)
      setSettings(saved.settings)
      setErrors({})
      setFailed(false)
      setMessage(runStarted ? 'Saved. Restart MVD to use these for new downloads.' : 'Saved.')
    } catch (error) {
      setErrors(error instanceof ApiError ? error.fields : {})
      setMessage(error instanceof Error ? error.message : String(error))
      setFailed(true)
    } finally {
      setBusy(false)
    }
  }

  const cookieChoices = [
    { value: 'all', label: 'Try every installed browser' },
    { value: 'off', label: 'Do not use cookies' },
    ...options.browsers.map(browser => ({ value: browser, label: browser })),
  ]
  if (cookieChoices.every(choice => choice.value !== settings.cookies)) {
    cookieChoices.push({ value: settings.cookies, label: settings.cookies })
  }

  return (
    <form className='settings' noValidate onSubmit={event => void submit(event)}>
      <FolderField id='outputDir' label='Download folder' value={settings.outputDir} error={errors.outputDir} onChange={value => update('outputDir', value)} />
      <Select id='videoQuality' label='Video quality' value={settings.videoQuality} choices={same(options.videoQualities)} error={errors.videoQuality} onChange={value => update('videoQuality', value)} />
      <Select id='audioQuality' label='Audio quality' value={settings.audioQuality} choices={same(options.audioQualities)} error={errors.audioQuality} onChange={value => update('audioQuality', value)} />
      <Select id='mergeOutputFormat' label='File format' value={settings.mergeOutputFormat} choices={same(options.mergeFormats)} error={errors.mergeOutputFormat} onChange={value => update('mergeOutputFormat', value)} />

      <div className='field'>
        <label htmlFor='outputTemplate'>File name template</label>
        <input id='outputTemplate' type='text' value={settings.outputTemplate} onChange={event => update('outputTemplate', event.target.value)} />
        {errors.outputTemplate && <span className='message error'>{errors.outputTemplate}</span>}
      </div>

      <div className='field'>
        <label htmlFor='maxConcurrentDownloads'>Videos at once</label>
        <input id='maxConcurrentDownloads' type='number' min={1} max={32} value={settings.maxConcurrentDownloads} onChange={event => number('maxConcurrentDownloads')(event.target.value)} />
        {errors.maxConcurrentDownloads && <span className='message error'>{errors.maxConcurrentDownloads}</span>}
      </div>
      <div className='field'>
        <label htmlFor='concurrentFragments'>Parts of one video at once (0 is off)</label>
        <input id='concurrentFragments' type='number' min={0} max={32} value={settings.concurrentFragments} onChange={event => number('concurrentFragments')(event.target.value)} />
        {errors.concurrentFragments && <span className='message error'>{errors.concurrentFragments}</span>}
      </div>

      <Select id='cookies' label='Browser sign-in (cookies)' value={settings.cookies} choices={cookieChoices} error={errors.cookies} onChange={value => update('cookies', value)} />

      <label className='check'>
        <input type='checkbox' checked={settings.downloadOfficialMusicVideo} onChange={event => update('downloadOfficialMusicVideo', event.target.checked)} />
        Prefer the official music video over auto-generated tracks
      </label>
      <label className='check'>
        <input type='checkbox' checked={settings.autoRetry} onChange={event => update('autoRetry', event.target.checked)} />
        Retry failed downloads automatically
      </label>
      <label className='check'>
        <input type='checkbox' checked={settings.createLogFile} onChange={event => update('createLogFile', event.target.checked)} />
        Keep a log file
      </label>
      {settings.createLogFile && (
        <FolderField id='logDir' label='Log folder' value={settings.logDir} error={errors.logDir} onChange={value => update('logDir', value)} />
      )}

      <div className='settings-actions'>
        <button type='submit' disabled={busy}>{busy ? 'Saving...' : 'Save settings'}</button>
        <span role='status' className={failed ? 'message error' : 'message'}>{message}</span>
      </div>
    </form>
  )
}
