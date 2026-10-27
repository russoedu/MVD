/**
 * The preferences as the program sends them over the mvd:prefs event while they are open, and what
 * the window sends back over mvd:prefs-intent. The shapes are libs/mvd-core/prefsform's Window and
 * WindowIntent; the two change together.
 */
export type PreferenceKind = 'choice' | 'number' | 'text' | 'folder' | 'template' | 'colour' | 'action'

export interface PreferenceOption {
  value:     string
  label:     string
  /** Colours shown beside the option, for a theme. */
  swatches?: string[]
}

/** A piece of the name pattern: a block, or text typed between blocks. */
export interface PatternUnit {
  text:   string
  block?: boolean
}

export interface PreferenceField {
  label:        string
  hint?:        string
  kind:         PreferenceKind
  value:        string
  options?:     PreferenceOption[]
  min?:         number
  max?:         number
  step?:        number
  /** What a number at 0 reads, when it is not "0". */
  off?:         string
  placeholder?: string
  /** A colour that may be left empty, for the terminal's own. */
  optional?:    boolean
  /** The name pattern's pieces, the blocks that can be added, and the name it gives a sample song. */
  units?:       PatternUnit[]
  blocks?:      string[]
  example?:     string
}

/** The colours of the terminal look, which MVD's own look draws the window in. */
export interface TerminalColours {
  background: string
  accent:     string
  focus:      string
  highlight:  string
  success:    string
  error:      string
  dim:        string
  text:       string
  selected:   string
}

export interface PreferencesSnapshot {
  title:    string
  /** What the first start says, until the person does something. */
  welcome?: string
  tabs:     string[]
  tab:      number
  /** The settings of the tab shown. */
  fields:   PreferenceField[]
  status?:  string
  bad?:     boolean
  /** True while a window of the program's (the folder chooser, a question) has the answer. */
  busy?:    boolean
  colours:  TerminalColours
}

export type PreferencesIntent =
  | { kind: 'prefs-tab', index: number } |
  { kind: 'prefs-set', field: number, text: string } |
  { kind: 'prefs-activate', field: number } |
  { kind: 'prefs-default-colour', field: number } |
  { kind: 'prefs-add-block', field: number, index: number } |
  { kind: 'prefs-add-text', field: number, text: string } |
  { kind: 'prefs-remove-unit', field: number, index: number } |
  { kind: 'prefs-save' } |
  { kind: 'prefs-close' }
