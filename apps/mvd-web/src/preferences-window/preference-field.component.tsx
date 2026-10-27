import { useState } from 'react'
import { Button, Group, Input, Label, ListBox, ListBoxItem, NumberField, Popover, Select, SelectValue, Text, TextField } from 'react-aria-components'
import { NamePatternField } from './name-pattern-field.component'
import type { PreferenceField as Field, PreferencesIntent } from './preferences-window.contract'

interface PreferenceFieldProps {
  field:    Field
  index:    number
  onIntent: (intent: PreferencesIntent) => void
}

/** The colours shown beside an option, such as a theme's. */
function Swatches ({ colours }: { colours?: string[] }) {
  if (!colours?.length) return null

  return (
    <span className='preferences-swatches' aria-hidden='true'>
      {[...new Set(colours)].map(colour => <span key={colour} style={{ background: colour }} />)}
    </span>
  )
}

/**
 * One setting, drawn by its kind: a drop-down for a choice, a box with arrows for a number, a text
 * box, a folder with Browse…, the name pattern, a colour with its swatch, or a button.
 */
export function PreferenceField ({ field, index, onIntent }: PreferenceFieldProps) {
  const set = (text: string): void => onIntent({ kind: 'prefs-set', field: index, text })
  const hint = field.hint && <Text slot='description' className='preferences-hint'>{field.hint}</Text>

  switch (field.kind) {
    case 'choice': {
      return (
        <Select className='preferences-field' selectedKey={field.value} onSelectionChange={key => set(String(key))}>
          <Label className='preferences-label'>{field.label}</Label>
          <Button className='classic-field preferences-select'>
            <SelectValue />
            <span aria-hidden='true' className='preferences-select-arrow'>▾</span>
          </Button>
          {hint}
          <Popover className='classic-menu-popover preferences-select-popover' placement='bottom start' offset={0}>
            <div className='classic-menu-panel'>
              <ListBox className='classic-menu' items={field.options ?? []}>
                {option => (
                  <ListBoxItem id={option.value} className='classic-menu-item' textValue={option.label}>
                    <span className='classic-menu-label'>{option.label}</span>
                    <Swatches colours={option.swatches} />
                  </ListBoxItem>
                )}
              </ListBox>
            </div>
          </Popover>
        </Select>
      )
    }
    case 'number': {
      return (
        <NumberField
          className='preferences-field'
          value={Number(field.value)}
          minValue={field.min}
          maxValue={field.max}
          step={field.step || 1}
          onChange={value => {
            if (!Number.isNaN(value)) set(String(value))
          }}
        >
          <Label className='preferences-label'>{field.label}</Label>
          <Group className='preferences-number'>
            <Button slot='decrement' className='classic-button preferences-step'>−</Button>
            <Input className='classic-field' />
            <Button slot='increment' className='classic-button preferences-step'>+</Button>
            {field.off && field.value === '0' && <span className='preferences-off'>{field.off}</span>}
          </Group>
          {hint}
        </NumberField>
      )
    }
    case 'template': {
      return (
        <div className='preferences-field'>
          <span className='preferences-label'>{field.label}</span>
          {field.hint && <span className='preferences-hint'>{field.hint}</span>}
          <NamePatternField
            field={field}
            onAdd={block => onIntent({ kind: 'prefs-add-block', field: index, index: block })}
            onText={text => onIntent({ kind: 'prefs-add-text', field: index, text })}
            onRemove={unit => onIntent({ kind: 'prefs-remove-unit', field: index, index: unit })}
          />
        </div>
      )
    }
    case 'action': {
      return (
        <div className='preferences-field'>
          <Button className='classic-button' onPress={() => onIntent({ kind: 'prefs-activate', field: index })}>{field.label}…</Button>
          {field.hint && <span className='preferences-hint'>{field.hint}</span>}
        </div>
      )
    }
    default: {
      // Drawn again when the program's value changes, so the box shows it.
      return <TypedField key={field.value} field={field} index={index} onIntent={onIntent} set={set} />
    }
  }
}

/**
 * A setting that is typed: a text, a folder (with Browse…) or a colour (with its swatch and
 * Default). What is typed is sent when it is finished: on Enter, or when the box is left.
 */
function TypedField ({ field, index, onIntent, set }: PreferenceFieldProps & { set: (text: string) => void }) {
  const [draft, setDraft] = useState(field.value)
  const commit = (): void => {
    if (draft !== field.value) set(draft)
  }
  const hint = field.hint && <Text slot='description' className='preferences-hint'>{field.hint}</Text>

  return (
    <TextField className='preferences-field' value={draft} onChange={setDraft} onBlur={commit} onKeyDown={event => { if (event.key === 'Enter') commit() }}>
      <Label className='preferences-label'>{field.label}</Label>
      <div className='preferences-row'>
        {field.kind === 'colour' && <span className='preferences-colour' style={{ background: field.value || 'transparent' }} aria-hidden='true' />}
        <Input className='classic-field' placeholder={field.placeholder} />
        {field.kind === 'folder' && (
          <Button className='classic-button preferences-small-button' onPress={() => onIntent({ kind: 'prefs-activate', field: index })}>Browse…</Button>
        )}
        {field.kind === 'colour' && (
          <Button className='classic-button preferences-small-button' onPress={() => onIntent({ kind: 'prefs-default-colour', field: index })}>Default</Button>
        )}
      </div>
      {hint}
    </TextField>
  )
}
