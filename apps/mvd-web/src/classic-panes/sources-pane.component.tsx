import { useRef, useState } from 'react'
import { ListBox, ListBoxItem, Menu, MenuItem, Popover } from 'react-aria-components'
import type { SourcesSnapshot } from '../workspace-snapshot'

/** What can be done to a source from its menu: a menu entry of the program, and whether it can now. */
export interface SourceAction {
  id:      string
  label:   string
  enabled: boolean
}

interface SourcesPaneProps {
  sources:   SourcesSnapshot
  onSelect:  (index: number) => void
  /** The entries of a source's own menu (a right click, or the menu key); none for no menu. */
  actions?:  SourceAction[]
  onAction?: (id: string) => void
}

/** Where a source's menu opens, in the page's own pixels. */
interface MenuPlace {
  x: number
  y: number
}

/** The page's zoom, which a bitmap theme sets to stay sharp: a click is in screen pixels. */
const pageZoom = (): number => Number(getComputedStyle(document.documentElement).getPropertyValue('--pixel-zoom')) || 1

/**
 * The sources of the list: platform addresses and searches by name, one row each. A right click on
 * one, or the menu key, chooses it and opens its menu: remove it, move it up or down.
 */
export function SourcesPane ({ sources, onSelect, actions = [], onAction }: SourcesPaneProps) {
  const selected = sources.rows[sources.selected]?.id
  const [menu, setMenu] = useState<MenuPlace | null>(null)
  const anchorRef = useRef<HTMLSpanElement>(null)

  const openMenu = (index: number, x: number, y: number): void => {
    if (actions.length === 0) return
    onSelect(index)
    const zoom = pageZoom()
    setMenu({ x: x / zoom, y: y / zoom })
  }

  return (
    <section className='classic-pane classic-sources' aria-label={sources.heading}>
      <h2 className='classic-pane-heading'>{sources.heading}</h2>
      <ListBox
        className='classic-pane-body classic-list'
        aria-label={sources.heading}
        selectionMode='single'
        disallowEmptySelection
        selectedKeys={selected ? [selected] : []}
        onSelectionChange={keys => {
          const key = [...keys][0]
          const index = sources.rows.findIndex(row => row.id === key)
          if (index !== -1) onSelect(index)
        }}
      >
        {sources.rows.map((row, index) => (
          <ListBoxItem
            key={row.id}
            id={row.id}
            className='classic-row'
            textValue={row.label}
            onContextMenu={event => {
              event.preventDefault()
              openMenu(index, event.clientX, event.clientY)
            }}
            onKeyDown={event => {
              if (event.key !== 'ContextMenu' && !(event.key === 'F10' && event.shiftKey)) return
              event.preventDefault()
              const box = event.currentTarget.getBoundingClientRect()
              openMenu(index, box.left + 12, box.bottom)
            }}
          >
            <span className='classic-source-label'>{row.label}</span>
            <span className='classic-source-detail' data-tone={row.tone}>{row.detail}</span>
          </ListBoxItem>
        ))}
      </ListBox>
      <span ref={anchorRef} className='classic-menu-anchor' style={{ left: menu?.x ?? 0, top: menu?.y ?? 0 }} />
      <Popover
        className='classic-menu-popover'
        triggerRef={anchorRef}
        placement='bottom start'
        offset={0}
        isOpen={menu !== null}
        onOpenChange={open => { if (!open) setMenu(null) }}
      >
        <div className='classic-menu-panel'>
          <Menu
            className='classic-menu'
            aria-label='Source'
            // React Aria's focus for the menu that opens, as a right click lights its first entry.
            // eslint-disable-next-line jsx-a11y/no-autofocus
            autoFocus='first'
            disabledKeys={actions.filter(action => !action.enabled).map(action => action.id)}
            onAction={id => {
              setMenu(null)
              onAction?.(String(id))
            }}
          >
            {actions.map(action => (
              <MenuItem key={action.id} id={action.id} className='classic-menu-item' textValue={action.label}>
                <span className='classic-menu-label'>{action.label}</span>
              </MenuItem>
            ))}
          </Menu>
        </div>
      </Popover>
    </section>
  )
}
