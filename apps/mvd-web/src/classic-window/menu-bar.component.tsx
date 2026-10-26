import { Fragment } from 'react'
import { Button, Keyboard, Menu, MenuItem, MenuTrigger, Popover, Separator, Toolbar } from 'react-aria-components'
import { shortcutLabel, type ChromeLayout } from '../classic-themes'
import type { MenuSnapshot } from '../workspace-snapshot'

interface MenuBarProps {
  menus:   MenuSnapshot[]
  layout:  ChromeLayout
  onEntry: (id: string) => void
}

/** The menu bar inside the window, with each menu's entries and their shortcuts written the family's way. */
export function MenuBar ({ menus, layout, onEntry }: MenuBarProps) {
  return (
    <Toolbar className='classic-menu-bar' aria-label='Menu bar'>
      {menus.map(menu => (
        <MenuTrigger key={menu.label}>
          <Button className='classic-menu-title'>{menu.label}</Button>
          <Popover className='classic-menu-popover' placement='bottom start' offset={0}>
            <Menu
              className='classic-menu'
              aria-label={menu.label}
              disabledKeys={menu.entries.filter(entry => !entry.enabled).map(entry => entry.id)}
              onAction={key => onEntry(String(key))}
            >
              {menu.entries.map(entry => (
                <Fragment key={entry.id}>
                  {entry.separatorBefore && <Separator className='classic-menu-separator' />}
                  <MenuItem id={entry.id} className='classic-menu-item' textValue={entry.label}>
                    <span className='classic-menu-label'>{entry.label}</span>
                    {entry.shortcut && <Keyboard className='classic-menu-shortcut'>{shortcutLabel(entry.shortcut, layout.shortcutStyle)}</Keyboard>}
                  </MenuItem>
                </Fragment>
              ))}
            </Menu>
          </Popover>
        </MenuTrigger>
      ))}
    </Toolbar>
  )
}
