import { Button, Dialog, Heading, Input, Label, Modal, ModalOverlay, Radio, RadioGroup, TextField } from 'react-aria-components'
import { orderDialogButtons, type ChromeLayout } from '../classic-themes'
import { TitleBar } from '../classic-window'
import type { AddMusicField, AddMusicMode, AddMusicSnapshot } from '../workspace-snapshot'

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
 * Windows it is a dialog with its own title bar; on Mac OS X it is a sheet under the title bar.
 */
export function AddMusicWindow ({ adding, layout, container, onMode, onText, onButton }: AddMusicWindowProps) {
  const sheet = layout.dialogStyle === 'sheet'
  const cancel = adding.buttons.find(button => button.role === 'cancel')

  return (
    <ModalOverlay
      className='classic-dialog-overlay'
      data-style={layout.dialogStyle}
      isOpen
      UNSTABLE_portalContainer={container}
      onOpenChange={open => { if (!open && cancel) onButton(cancel.id) }}
    >
      <Modal className={sheet ? 'classic-sheet' : 'classic-dialog'}>
        <Dialog className='classic-dialog-content' aria-label={sheet ? adding.heading : undefined}>
          {!sheet && (
            <TitleBar
              title={<Heading slot='title' className='classic-title'>{adding.heading}</Heading>}
              layout={layout}
              controls={cancel ? ['close'] : []}
              onWindow={action => { if (action === 'close' && cancel) onButton(cancel.id) }}
            />
          )}
          <form
            className='classic-dialog-body'
            onSubmit={event => {
              event.preventDefault()
              const primary = adding.buttons.find(button => button.role === 'default')
              if (primary?.enabled) onButton(primary.id)
            }}
          >
            {sheet && <Heading slot='title' className='classic-sheet-heading'>{adding.heading}</Heading>}
            <RadioGroup
              className='classic-radio-group'
              aria-label='What to add'
              orientation='horizontal'
              value={adding.mode}
              onChange={mode => onMode(mode as AddMusicMode)}
            >
              {adding.modes.map(({ mode, label }) => (
                <Radio key={mode} value={mode} className='classic-radio'>
                  <span className='classic-radio-dot' aria-hidden='true' />
                  {label}
                </Radio>
              ))}
            </RadioGroup>
            {adding.fields.map(({ field, label, text }) => (
              <TextField key={field} className='classic-text-field' value={text} onChange={value => onText(field, value)}>
                <Label className='classic-field-label'>{label}</Label>
                <Input className='classic-field' />
              </TextField>
            ))}
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
                  {button.label}
                </Button>
              ))}
            </div>
          </form>
        </Dialog>
      </Modal>
    </ModalOverlay>
  )
}
