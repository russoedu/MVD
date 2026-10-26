import { Fragment, useEffect, useRef, useState, type KeyboardEvent as ReactKeyboardEvent } from 'react'
import { Button, Keyboard, Menu, MenuItem, MenuTrigger, Popover, Separator, Toolbar } from 'react-aria-components'
import { isMnemonic, pressedLetter, shortcutLabel, type ChromeLayout } from '../classic-themes'
import type { MenuSnapshot } from '../workspace-snapshot'
import { MnemonicLabel } from './mnemonic-label.component'

interface MenuBarProps {
  menus:   MenuSnapshot[]
  layout:  ChromeLayout
  onEntry: (id: string) => void
  /** True while a dialog covers the window: its keyboard letters are the dialog's, not the menus'. */
  blocked: boolean
}

/** Which menu is open, and whether the keyboard opened it, so its first entry is lit as Windows does. */
interface OpenMenu {
  label: string
  byKey: boolean
}

/**
 * The menu bar inside the window, with each menu's entries and their shortcuts written the
 * family's way. On Windows, Alt and a menu's letter opens it, a letter in an open menu picks the
 * entry, and Alt or F10 alone puts the focus on the bar. Left and right move between menus.
 */
export function MenuBar ({ menus, layout, onEntry, blocked }: MenuBarProps) {
  const [open, setOpen] = useState<OpenMenu | null>(null)
  const barRef = useRef<HTMLDivElement>(null)
  const letters = layout.mnemonics !== 'none'

  const openByLetter = (letter: string | null): boolean => {
    const menu = menus.find(each => isMnemonic(letter, each.mnemonic))
    if (menu) setOpen({ label: menu.label, byKey: true })

    return Boolean(menu)
  }

  useEffect(() => {
    if (!letters || blocked) return
    let altAlone = false

    const focusBar = (): void => {
      const titles = barRef.current?.querySelectorAll<HTMLElement>('.classic-menu-title')
      if (!titles?.length) return
      if (barRef.current?.contains(document.activeElement)) (document.activeElement as HTMLElement).blur()
      else titles[0].focus()
    }
    const onKeyDown = (event: KeyboardEvent): void => {
      altAlone = event.key === 'Alt'
      if (altAlone || event.defaultPrevented) return
      if (event.key === 'F10' && !event.altKey && !event.ctrlKey && !event.shiftKey && !event.metaKey) {
        event.preventDefault()
        focusBar()
      } else if (event.altKey && !event.ctrlKey && !event.metaKey && openByLetter(pressedLetter(event))) {
        event.preventDefault()
      }
    }
    const onKeyUp = (event: KeyboardEvent): void => {
      if (!altAlone || event.key !== 'Alt') return
      altAlone = false
      event.preventDefault()
      focusBar()
    }
    const onPointer = (): void => { altAlone = false }
    document.addEventListener('keydown', onKeyDown)
    document.addEventListener('keyup', onKeyUp)
    document.addEventListener('pointerdown', onPointer, { capture: true })

    return () => {
      document.removeEventListener('keydown', onKeyDown)
      document.removeEventListener('keyup', onKeyUp)
      document.removeEventListener('pointerdown', onPointer, { capture: true })
    }
  })

  const onMenuKey = (event: ReactKeyboardEvent, menu: MenuSnapshot): void => {
    if (event.ctrlKey || event.metaKey) return
    const step = { ArrowLeft: -1, ArrowRight: 1 }[event.key]
    let handled = false
    if (step && !event.altKey) {
      const next = menus[(menus.indexOf(menu) + step + menus.length) % menus.length]
      setOpen({ label: next.label, byKey: true })
      handled = true
    } else if (letters && event.altKey) {
      handled = openByLetter(pressedLetter(event))
    } else if (letters) {
      const entry = menu.entries.find(each => each.enabled && isMnemonic(pressedLetter(event), each.mnemonic))
      if (entry) {
        setOpen(null)
        onEntry(entry.id)
        handled = true
      }
    }
    if (!handled) return
    event.preventDefault()
    event.stopPropagation()
  }

  return (
    <Toolbar ref={barRef} className='classic-menu-bar' aria-label='Menu bar'>
      {menus.map(menu => (
        <MenuTrigger
          key={menu.label}
          isOpen={open?.label === menu.label}
          onOpenChange={isOpen => setOpen(isOpen ? { label: menu.label, byKey: false } : null)}
        >
          <Button className='classic-menu-title'><MnemonicLabel label={menu.label} mnemonic={menu.mnemonic} /></Button>
          <Popover className='classic-menu-popover' placement='bottom start' offset={0}>
            <div className='classic-menu-panel' onKeyDownCapture={event => onMenuKey(event, menu)}>
              <Menu
                className='classic-menu'
                aria-label={menu.label}
                // React Aria's focus for the menu that opens, not the page's autofocus: the keyboard
                // lights the first entry, as Windows does, and the pointer lights none.
                // eslint-disable-next-line jsx-a11y/no-autofocus
                autoFocus={open?.byKey ? 'first' : undefined}
                disabledKeys={menu.entries.filter(entry => !entry.enabled).map(entry => entry.id)}
                onAction={key => onEntry(String(key))}
              >
                {menu.entries.map(entry => (
                  <Fragment key={entry.id}>
                    {entry.separatorBefore && <Separator className='classic-menu-separator' />}
                    <MenuItem id={entry.id} className='classic-menu-item' textValue={entry.label}>
                      <span className='classic-menu-label'><MnemonicLabel label={entry.label} mnemonic={entry.mnemonic} /></span>
                      {entry.shortcut && <Keyboard className='classic-menu-shortcut'>{shortcutLabel(entry.shortcut, layout.shortcutStyle)}</Keyboard>}
                    </MenuItem>
                  </Fragment>
                ))}
              </Menu>
            </div>
          </Popover>
        </MenuTrigger>
      ))}
    </Toolbar>
  )
}
