import { ListBox, ListBoxItem } from 'react-aria-components'
import type { SourcesSnapshot } from '../workspace-snapshot'

interface SourcesPaneProps {
  sources:  SourcesSnapshot
  onSelect: (index: number) => void
}

/** The sources of the list: platform addresses and searches by name, one row each. */
export function SourcesPane ({ sources, onSelect }: SourcesPaneProps) {
  const selected = sources.rows[sources.selected]?.id

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
        {sources.rows.map(row => (
          <ListBoxItem key={row.id} id={row.id} className='classic-row' textValue={row.label}>
            <span className='classic-source-label'>{row.label}</span>
            <span className='classic-source-detail' data-tone={row.tone}>{row.detail}</span>
          </ListBoxItem>
        ))}
      </ListBox>
    </section>
  )
}
