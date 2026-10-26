import { Button, Toolbar } from 'react-aria-components'
import type { ButtonSnapshot } from '../workspace-snapshot'

interface ActionsToolbarProps {
  buttons:  ButtonSnapshot[]
  onButton: (id: string) => void
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
            {button.label}
          </Button>
        </span>
      ))}
    </Toolbar>
  )
}
