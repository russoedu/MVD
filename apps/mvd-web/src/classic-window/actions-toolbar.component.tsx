import { Button, Toolbar } from 'react-aria-components'
import type { ButtonSnapshot } from '../workspace-snapshot'

interface ActionsToolbarProps {
  buttons:  ButtonSnapshot[]
  onButton: (id: string) => void
}

/** A button's words; the mark in front (► ■ ✓ +) is drawn in a font that has it. */
function ButtonLabel ({ label }: { label: string }) {
  const mark = /^([►■✓+]) (.*)$/.exec(label)
  if (!mark) return <>{label}</>

  return <><span className='classic-glyph'>{mark[1]}</span> {mark[2]}</>
}

/** The row of buttons under the menu bar: what can be done next. A disabled one says why when pointed at. */
export function ActionsToolbar ({ buttons, onButton }: ActionsToolbarProps) {
  return (
    <Toolbar className='classic-tool-bar' aria-label='Actions'>
      {buttons.map(button => (
        <span key={button.id} title={button.enabled ? undefined : button.disabledReason}>
          <Button
            className='classic-button'
            data-role={button.role}
            isDisabled={!button.enabled}
            onPress={() => onButton(button.id)}
          >
            <ButtonLabel label={button.label} />
          </Button>
        </span>
      ))}
    </Toolbar>
  )
}
