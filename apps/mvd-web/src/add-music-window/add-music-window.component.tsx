import { Button, Dialog, Heading, Input, Label, Modal, ModalOverlay, Radio, RadioGroup, TextField } from 'react-aria-components'
import { useEffect, useRef } from 'react'
import { isMnemonic, orderDialogButtons, pressedLetter, type ChromeLayout } from '../classic-themes'
import { MnemonicLabel, TitleBar, useEchoedText } from '../classic-window'
import type { AddMusicField, AddMusicMode, AddMusicSnapshot } from '../workspace-snapshot'

interface AddMusicFieldProps {
  field:  AddMusicSnapshot['fields'][number]
  onText: (field: AddMusicField, text: string) => void
}

/** One field of the dialog: it shows what is typed at once, while the program keeps the text. */
function AddMusicTextField ({ field: { field, label, mnemonic, text }, onText }: AddMusicFieldProps) {
  const [shown, write] = useEchoedText(text, next => onText(field, next))

  return (
    <TextField className='classic-text-field' data-field={field} value={shown} onChange={write}>
      <Label className='classic-field-label'><MnemonicLabel label={label} mnemonic={mnemonic} /></Label>
      <Input className='classic-field' />
    </TextField>
  )
}

interface AddMusicWindowProps {
  adding:    AddMusicSnapshot
  layout:    ChromeLayout
  /** The element inside the classic window the dialog is drawn in, so a sheet hangs from its title bar. */
  container: Element
  onMode:    (mode: AddMusicMode) => void
  onText:    (field: AddMusicField, text: string) => void
  onButton:  (id: string) => void
}

/**
 * The window for adding music: a platform address, or a search by artist, album and song. On
 * Windows it is a dialog with its own title bar; on a classic Mac a plain framed box with its
 * heading inside; on Mac OS X a sheet under the title bar.
 */
export function AddMusicWindow ({ adding, layout, container, onMode, onText, onButton }: AddMusicWindowProps) {
  const sheet = layout.dialogStyle === 'sheet'
  const titled = layout.dialogStyle === 'window'
  const cancel = adding.buttons.find(button => button.role === 'cancel')
  const primary = adding.buttons.find(button => button.role === 'default')

  const formRef = useRef<HTMLFormElement>(null)

  // The cursor is in the first field when the dialog opens and when another mode changes the fields.
  useEffect(() => {
    formRef.current?.querySelector<HTMLInputElement>('input[type=text]')?.focus()
  }, [adding.mode])

  // Windows: Alt and a letter reaches the choice, field or button with that letter; so does the
  // letter alone when the focus is not in a text field. The dialog is modal, so every key is its.
  useEffect(() => {
    if (layout.mnemonics === 'none') return
    const onKeyDown = (event: KeyboardEvent): void => {
      const typing = event.target instanceof HTMLInputElement && event.target.type === 'text'
      if (event.defaultPrevented || event.ctrlKey || event.metaKey || (typing && !event.altKey)) return
      const letter = pressedLetter(event)
      const mode = adding.modes.find(each => isMnemonic(letter, each.mnemonic))
      const field = adding.fields.find(each => isMnemonic(letter, each.mnemonic))
      const button = adding.buttons.find(each => each.enabled && isMnemonic(letter, each.mnemonic))
      if (mode) onMode(mode.mode)
      else if (field) formRef.current?.querySelector<HTMLInputElement>(`[data-field='${CSS.escape(field.field)}'] input`)?.focus()
      else if (button) onButton(button.id)
      else return
      event.preventDefault()
    }
    document.addEventListener('keydown', onKeyDown)

    return () => document.removeEventListener('keydown', onKeyDown)
  })

  return (
    <ModalOverlay
      className='classic-dialog-overlay'
      data-style={layout.dialogStyle}
      isOpen
      UNSTABLE_portalContainer={container}
      onOpenChange={open => { if (!open && cancel) onButton(cancel.id) }}
    >
      <Modal className={sheet ? 'classic-sheet' : 'classic-dialog'} data-style={layout.dialogStyle}>
        <Dialog className='classic-dialog-content' aria-label={titled ? undefined : adding.heading}>
          {titled && (
            <TitleBar
              title={<Heading slot='title' className='classic-title'>{adding.heading}</Heading>}
              layout={layout}
              controls={cancel ? ['close'] : []}
              onWindow={action => { if (action === 'close' && cancel) onButton(cancel.id) }}
            />
          )}
          <form
            ref={formRef}
            className='classic-dialog-body'
            onSubmit={event => {
              event.preventDefault()
              if (primary?.enabled) onButton(primary.id)
            }}
          >
            {!titled && <Heading slot='title' className='classic-sheet-heading'>{adding.heading}</Heading>}
            <RadioGroup
              className='classic-radio-group'
              aria-label='What to add'
              orientation='horizontal'
              value={adding.mode}
              onChange={mode => onMode(mode as AddMusicMode)}
            >
              {adding.modes.map(({ mode, label, mnemonic }) => (
                <Radio key={mode} value={mode} className='classic-radio'>
                  <span className='classic-radio-dot' aria-hidden='true' />
                  <MnemonicLabel label={label} mnemonic={mnemonic} />
                </Radio>
              ))}
            </RadioGroup>
            {adding.fields.map(each => <AddMusicTextField key={each.field} field={each} onText={onText} />)}
            <ul className='classic-dialog-help'>
              {adding.help.map(line => <li key={line}>{line}</li>)}
            </ul>
            {adding.problem && <p className='classic-dialog-problem' role='alert'>{adding.problem}</p>}
            <div className='classic-dialog-buttons' data-side={layout.defaultButton === 'first' ? 'windows' : 'mac'}>
              {orderDialogButtons(adding.buttons, layout).map(button => (
                <Button
                  key={button.id}
                  className='classic-button'
                  data-role={button.role}
                  type={button.role === 'default' ? 'submit' : 'button'}
                  isDisabled={!button.enabled}
                  onPress={button.role === 'default' ? undefined : () => onButton(button.id)}
                >
                  <MnemonicLabel label={button.label} mnemonic={button.mnemonic} />
                </Button>
              ))}
            </div>
          </form>
        </Dialog>
      </Modal>
    </ModalOverlay>
  )
}
