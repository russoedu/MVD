/**
 * What MVD's main window shows, as the program will send it to a classic theme, and what the
 * theme sends back. It describes roles, not places: a theme decides where a default button
 * or a shortcut goes, so the same snapshot draws as Windows or as a Mac.
 *
 * The program does not send it yet (the JSON channel of #345 comes after #344); the theme
 * preview draws a sample of it.
 */
export interface WorkspaceSnapshot {
  title:    string
  menus:    MenuSnapshot[]
  toolbar:  ButtonSnapshot[]
  sources:  SourcesSnapshot
  songs:    SongsSnapshot
  details:  DetailsSnapshot
  status:   StatusSnapshot
  activity: Activity
  /** The window for adding music, when it is open. */
  adding?:  AddMusicSnapshot
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

export interface MenuSnapshot {
  label:     string
  mnemonic?: Mnemonic
  entries:   MenuEntrySnapshot[]
}

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

export interface SourceRow {
  id:     string
  label:  string
  detail: string
  tone:   Tone
}

export interface SourcesSnapshot {
  heading:  string
  rows:     SourceRow[]
  selected: number
}

/** Progress from 0 to 1, 'unknown' while it cannot be measured, or null when there is none. */
export type Progress = number | 'unknown' | null

export interface SongRow {
  id:       string
  ticked:   boolean
  artist:   string
  title:    string
  progress: Progress
  status:   string
  tone:     Tone
}

export interface SongsSnapshot {
  heading:  string
  rows:     SongRow[]
  selected: number
}

export interface DetailsSnapshot {
  heading: string
  lines:   string[]
}

export interface KeyHint {
  shortcut: Shortcut
  label:    string
}

export interface StatusSnapshot {
  summary: string
  hints:   KeyHint[]
}

export type AddMusicMode = 'url' | 'search'
export type AddMusicField = 'url' | 'artist' | 'album' | 'song'

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
  { kind: 'adding-button', id: string }
