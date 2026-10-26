import { Cell, Checkbox, Column, Row, Table, TableBody, TableHeader } from 'react-aria-components'
import type { SongsSnapshot } from '../workspace-snapshot'
import { ProgressBar } from './progress-bar.component'

interface SongsTableProps {
  songs:    SongsSnapshot
  onSelect: (index: number) => void
  onTick:   (index: number, ticked: boolean) => void
}

/** The songs of the chosen source: ticked to download, with their progress and status. */
export function SongsTable ({ songs, onSelect, onTick }: SongsTableProps) {
  const selected = songs.rows[songs.selected]?.id

  return (
    <section className='classic-pane classic-songs' aria-label={songs.heading}>
      <h2 className='classic-pane-heading'>{songs.heading}</h2>
      <div className='classic-pane-body'>
        <Table
          className='classic-table'
          aria-label={songs.heading}
          selectionMode='single'
          selectionBehavior='replace'
          disallowEmptySelection
          selectedKeys={selected ? [selected] : []}
          onSelectionChange={keys => {
            if (keys === 'all') return
            const index = songs.rows.findIndex(row => row.id === [...keys][0])
            if (index !== -1) onSelect(index)
          }}
        >
          <TableHeader className='classic-table-header'>
            <Column className='classic-column classic-column-tick' aria-label='Download'> </Column>
            <Column className='classic-column' isRowHeader>Artist</Column>
            <Column className='classic-column'>Song</Column>
            <Column className='classic-column classic-column-progress'>Progress</Column>
            <Column className='classic-column'>Status</Column>
          </TableHeader>
          <TableBody className='classic-table-body'>
            {songs.rows.map((row, index) => (
              <Row key={row.id} id={row.id} className='classic-row'>
                <Cell className='classic-cell'>
                  <Checkbox
                    className='classic-check'
                    aria-label={`Download ${row.title}`}
                    isSelected={row.ticked}
                    onChange={ticked => onTick(index, ticked)}
                  >
                    <span className='classic-check-box' aria-hidden='true' />
                  </Checkbox>
                </Cell>
                <Cell className='classic-cell'>{row.artist}</Cell>
                <Cell className='classic-cell'>{row.title}</Cell>
                <Cell className='classic-cell'><ProgressBar progress={row.progress} label={`${row.title} progress`} /></Cell>
                <Cell className='classic-cell' data-tone={row.tone}>{row.status}</Cell>
              </Row>
            ))}
          </TableBody>
        </Table>
      </div>
    </section>
  )
}
