import { Button, Dialog, Heading, Input, Label, Modal, ModalOverlay, TextField } from 'react-aria-components'
import { useEffect, useRef } from 'react'
import { orderDialogButtons, type ChromeLayout } from '../classic-themes'
import { TitleBar, useEchoedText } from '../classic-window'
import type { ButtonSnapshot, QuestionSnapshot } from '../workspace-snapshot'

interface QuestionWindowProps {
  question:  QuestionSnapshot
  layout:    ChromeLayout
  /** The element inside the classic window the dialog is drawn in, so a sheet hangs from its title bar. */
  container: Element
  onButton:  (id: string) => void
  onText:    (text: string) => void
}

/** The button Esc and the close box press: Cancel, else the default one (Keep downloading, OK). */
function escapeButton (buttons: ButtonSnapshot[]): ButtonSnapshot | undefined {
  return buttons.find(button => button.role === 'cancel') ?? buttons.find(button => button.role === 'default')
}

/**
 * A question or note over the window, as the theme's system drew one: an alert with its mark (a
 * caution, a note, a question), the lines it says, a table of keys, a line to write in, and the
 * buttons where the system puts them, the default one pressed by Return.
 */
export function QuestionWindow ({ question, layout, container, onButton, onText }: QuestionWindowProps) {
  const sheet = layout.dialogStyle === 'sheet'
  const titled = layout.dialogStyle === 'window'
  const escape = escapeButton(question.buttons)
  const fieldRef = useRef<HTMLInputElement>(null)

  // A line to write has the cursor as soon as the question opens.
  useEffect(() => { fieldRef.current?.focus() }, [question.id])

  const [shown, write] = useEchoedText(question.field?.text ?? '', onText)
  const primary = question.buttons.find(button => button.role === 'default')

  return (
    <ModalOverlay
      className='classic-dialog-overlay'
      data-style={layout.dialogStyle}
      isOpen
      UNSTABLE_portalContainer={container}
      onOpenChange={open => { if (!open && escape) onButton(escape.id) }}
    >
      <Modal className={sheet ? 'classic-sheet classic-question' : 'classic-dialog classic-question'} data-style={layout.dialogStyle}>
        <Dialog className='classic-dialog-content' role={question.tone === 'caution' ? 'alertdialog' : 'dialog'} aria-label={titled ? undefined : question.title}>
          {titled && (
            <TitleBar
              title={<Heading slot='title' className='classic-title'>{question.title}</Heading>}
              layout={layout}
              controls={escape ? ['close'] : []}
              onWindow={() => { if (escape) onButton(escape.id) }}
            />
          )}
          <form
            className='classic-dialog-body'
            onSubmit={event => {
              event.preventDefault()
              if (primary?.enabled) onButton(primary.id)
            }}
          >
            <div className='classic-question-body'>
              <span className='classic-question-mark' data-tone={question.tone} aria-hidden='true' />
              <div className='classic-question-words'>
                {!titled && <Heading slot='title' className='classic-sheet-heading'>{question.title}</Heading>}
                {question.lines?.map((line, at) => <p key={`${at}-${line}`} className='classic-question-line'>{line}</p>)}
                {question.sections?.map(section => (
                  <table key={section.heading} className='classic-question-table'>
                    <caption>{section.heading}</caption>
                    <tbody>
                      {section.rows.map(row => (
                        <tr key={row.keys}>
                          <th scope='row'><kbd>{row.keys}</kbd></th>
                          <td>{row.does}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                ))}
                {question.field && (
                  <TextField className='classic-question-field' value={shown} onChange={write}>
                    <Label className='classic-question-label'>{question.field.label}</Label>
                    <Input ref={fieldRef} className='classic-field' />
                  </TextField>
                )}
              </div>
            </div>
            <div className='classic-dialog-buttons' data-side={layout.defaultButton === 'first' ? 'windows' : 'mac'}>
              {orderDialogButtons(question.buttons, layout).map(button => (
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
