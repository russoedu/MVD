/**
 * What MVD's main window shows, as the program will send it to a classic theme, and what the
 * theme sends back. It describes roles, not places: a theme decides where a default button
 * or a shortcut goes, so the same snapshot draws as Windows or as a Mac.
 *
 * The program sends it over the mvd:snapshot event (libs/mvd-core/workspace/windowsnapshot has
 * the same shapes) while a classic window style is chosen; the theme preview draws a sample.
 */
export interface WorkspaceSnapshot {
  title:     string
  menus:     MenuSnapshot[]
  toolbar:   ButtonSnapshot[]
  sources:   SourcesSnapshot
  songs:     SongsSnapshot
  details:   DetailsSnapshot
  status:    StatusSnapshot
  activity:  Activity
  /** The window for adding music, when it is open. */
  adding?:   AddMusicSnapshot
  /** The question or note open over the window: quitting, a line to write, About, the shortcuts. */
  question?: QuestionSnapshot
  /**
   * A screen only the terminal look draws so far (the preferences, a question...), when one is
   * open over the window: the theme shows the terminal for it, under this title.
   */
  covered?:  { title: string }
}

/**
 * What MVD is doing, for the theme's pointer: nothing; working while the user carries on (the
 * arrow with a small hourglass on Windows); or keeping the user waiting (the hourglass, the ring
 * or the spinning wheel).
 */
export type Activity = 'idle' | 'background' | 'waiting'

/**
 * The letter Windows underlines in a label and reaches with Alt and the letter, written as it
 * appears in the label; the first time it appears, in either case, is the one underlined. Mac
 * themes have no such letters and ignore it.
 */
export type Mnemonic = string

/** A shortcut as keys, so each theme writes it its own way (Ctrl+L or ⌘L). */
export interface Shortcut {
  ctrl?:  boolean
  shift?: boolean
  alt?:   boolean
  key:    string
}

/** One menu of the menu bar: its label and its entries. */
export interface MenuSnapshot {
  label:     string
  mnemonic?: Mnemonic
  entries:   MenuEntrySnapshot[]
}

/** One entry of a menu: what it says, its shortcut and whether it can be used. */
export interface MenuEntrySnapshot {
  id:               string
  label:            string
  mnemonic?:        Mnemonic
  shortcut?:        Shortcut
  enabled:          boolean
  /** Draws a separator line above the entry. */
  separatorBefore?: boolean
}

/** What a button means in its group: a theme places the default and the cancel buttons. */
export type ButtonRole = 'default' | 'cancel' | 'other'

/** One button of a dialog or of the window: its label, its role and whether it can be used. */
export interface ButtonSnapshot {
  id:              string
  label:           string
  mnemonic?:       Mnemonic
  role:            ButtonRole
  enabled:         boolean
  /** Why it cannot be used, when it cannot. */
  disabledReason?: string
}

/** How a line reads: a theme picks the colour. */
export type Tone = 'normal' | 'dim' | 'good' | 'bad' | 'busy'

/** One row of the list of sources: a link or a search, with its state. */
export interface SourceRow {
  id:     string
  label:  string
  detail: string
  tone:   Tone
}

/** The list of sources: its heading, its rows and the selected row. */
export interface SourcesSnapshot {
  heading:  string
  rows:     SourceRow[]
  selected: number
}

/** Progress from 0 to 1, 'unknown' while it cannot be measured, or null when there is none. */
export type Progress = number | 'unknown' | null

/** One row of the list of songs: the song, its progress and its status. */
export interface SongRow {
  id:       string
  ticked:   boolean
  artist:   string
  title:    string
  progress: Progress
  status:   string
  tone:     Tone
}

/** The list of songs: its heading, its rows and the selected row. */
export interface SongsSnapshot {
  heading:  string
  rows:     SongRow[]
  selected: number
}

/** The details of the selected song: a heading and its lines. */
export interface DetailsSnapshot {
  heading: string
  lines:   string[]
}

/** One hint of the status line: a shortcut and what it does. */
export interface KeyHint {
  shortcut: Shortcut
  label:    string
}

/** The status line: a summary and the hints for the keys that work now. */
export interface StatusSnapshot {
  summary: string
  hints:   KeyHint[]
}

/** How music is added: by the address of a link or by a search. */
export type AddMusicMode = 'url' | 'search'
/** A field of the Add music window. */
export type AddMusicField = 'url' | 'artist' | 'album' | 'song'

/** The Add music window: its mode, the fields of that mode, its help, any problem and its buttons. */
export interface AddMusicSnapshot {
  heading:  string
  mode:     AddMusicMode
  modes:    { mode: AddMusicMode, label: string, mnemonic?: Mnemonic }[]
  /** The fields of the chosen mode, in order, with their text. */
  fields:   { field: AddMusicField, label: string, mnemonic?: Mnemonic, text: string }[]
  /** What the chosen mode accepts, one line each. */
  help:     string[]
  /** What is wrong with what was written, when something is. */
  problem?: string
  buttons:  ButtonSnapshot[]
}

/** The mark a dialog wears: a caution for a choice that loses something, a note, a question. */
export type QuestionTone = 'caution' | 'note' | 'ask'

/**
 * A question or note over the window, which a theme draws as its system's dialog. A button sends
 * back its id; the field, what is written in it.
 */
export interface QuestionSnapshot {
  id:        'quit' | 'ask' | 'about' | 'shortcuts'
  title:     string
  tone:      QuestionTone
  lines?:    string[]
  /** A table of keys and what they do, under headings. */
  sections?: { heading: string, rows: { keys: string, does: string }[] }[]
  /** The line the question asks to write in. */
  field?:    { label: string, text: string }
  buttons:   ButtonSnapshot[]
}

/** What the user did in a theme, for the program to act on. */
export type WorkspaceIntent =
  | { kind: 'menu', id: string } |
  { kind: 'button', id: string } |
  { kind: 'select-source', index: number } |
  { kind: 'select-song', index: number } |
  { kind: 'tick-song', index: number, ticked: boolean } |
  { kind: 'window', action: 'minimise' | 'maximise' | 'close' } |
  { kind: 'adding-mode', mode: AddMusicMode } |
  { kind: 'adding-text', field: AddMusicField, text: string } |
  { kind: 'adding-button', id: string } |
  { kind: 'question-button', id: string } |
  { kind: 'question-text', text: string } |
  /** A key the window does not use itself, such as a letter that accepts a song. */
  { kind: 'key', key: string }
