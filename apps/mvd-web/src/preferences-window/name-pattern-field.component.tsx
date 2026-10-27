import { useState } from 'react'
import { Button, Input, TextField } from 'react-aria-components'
import type { PatternUnit, PreferenceField } from './preferences-window.contract'

interface NamePatternFieldProps {
  field:    PreferenceField
  onAdd:    (block: number) => void
  onText:   (text: string) => void
  onRemove: (unit: number) => void
}

/** A run of the pattern: a block, or the text between two blocks, with where it starts. */
interface Piece {
  unit:  PatternUnit
  start: number
  text:  string
}

/** The pattern's units with the text between blocks put together, as it reads. */
function piecesOf (units: PatternUnit[]): Piece[] {
  const pieces: Piece[] = []
  let last: Piece | undefined
  for (const [at, unit] of units.entries()) {
    if (last && !unit.block && !last.unit.block) {
      last.text += unit.text
    } else {
      last = { unit, start: at, text: unit.text }
      pieces.push(last)
    }
  }

  return pieces
}

/**
 * The name pattern of the files: its blocks, each with a button to take it out, the text between
 * them, the blocks that can be added at its end, a box to add text, and the name it gives a sample
 * song.
 */
export function NamePatternField ({ field, onAdd, onText, onRemove }: NamePatternFieldProps) {
  const [typed, setTyped] = useState('')
  const pieces = piecesOf(field.units ?? [])

  return (
    <div className='preferences-pattern'>
      <div className='preferences-pattern-units' aria-label={`${field.label}: ${field.example ?? ''}`}>
        {pieces.map(piece => piece.unit.block
          ? (
              <span key={piece.start} className='preferences-pattern-block'>
                {piece.text}
                <Button className='preferences-pattern-remove' aria-label={`Take out ${piece.text}`} onPress={() => onRemove(piece.start)}>×</Button>
              </span>
            )
          : <span key={piece.start} className='preferences-pattern-text'>{piece.text}</span>)}
        <span className='preferences-pattern-text'>.ext</span>
      </div>
      <div className='preferences-pattern-palette'>
        {(field.blocks ?? []).map((block, index) => (
          <Button key={block} className='classic-button preferences-small-button' onPress={() => onAdd(index)}>+ {block}</Button>
        ))}
        <TextField className='preferences-pattern-typed' aria-label='Text to add' value={typed} onChange={setTyped}>
          <Input className='classic-field' placeholder='text' />
        </TextField>
        <Button
          className='classic-button preferences-small-button'
          isDisabled={typed === ''}
          onPress={() => {
            onText(typed)
            setTyped('')
          }}
        >
          + Text
        </Button>
      </div>
      {field.example && <p className='preferences-example'>Example: {field.example}</p>}
    </div>
  )
}
