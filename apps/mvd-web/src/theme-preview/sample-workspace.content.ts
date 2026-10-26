import type { AddMusicSnapshot, MenuEntrySnapshot, Shortcut, WorkspaceSnapshot } from '../workspace-snapshot'

const ctrl = (key: string, shift = false): Shortcut => ({ ctrl: true, shift, key })

const entry = (id: string, label: string, shortcut?: Shortcut, more: Partial<MenuEntrySnapshot> = {}): MenuEntrySnapshot =>
  ({ id, label, shortcut, enabled: true, ...more })

/** The window for adding music, as it opens: the platform address chosen, nothing written yet. */
export const sampleAdding: AddMusicSnapshot = {
  heading: 'Add music',
  mode:    'url',
  modes:   [{ mode: 'url', label: 'Platform URL' }, { mode: 'search', label: 'Music search' }],
  fields:  [{ field: 'url', label: 'URL', text: '' }],
  help:    [
    'YouTube playlist or video',
    'Spotify playlist or album',
    'Apple Music playlist or album',
    'Several at once, separated by spaces or new lines',
  ],
  buttons: [
    { id: 'add', label: 'Add', role: 'default', enabled: true },
    { id: 'cancel', label: 'Cancel', role: 'cancel', enabled: true },
  ],
}

/** The search fields and what they find, shown when Music search is chosen. */
export const sampleSearch: Pick<AddMusicSnapshot, 'fields' | 'help'> = {
  fields: [
    { field: 'artist', label: 'Artist', text: '' },
    { field: 'album', label: 'Album', text: '' },
    { field: 'song', label: 'Song', text: '' },
  ],
  help: [
    'An artist alone finds every song of the artist',
    'An artist and an album find the songs of the album',
    'An artist and a song find that song',
  ],
}

/** A list part-way through a run, to draw the classic themes with until the program sends its own. */
export const sampleWorkspace: WorkspaceSnapshot = {
  title: 'MVD - Music Video Downloader - my-list.mvd',
  menus: [
    {
      label:   'File',
      entries: [
        entry('new', 'New list', ctrl('N')),
        entry('open', 'Open list…', ctrl('O')),
        entry('save', 'Save list', ctrl('S')),
        entry('save-as', 'Save list as…', ctrl('S', true)),
        entry('preferences', 'Preferences…', ctrl(','), { separatorBefore: true }),
        entry('quit', 'Quit', ctrl('Q'), { separatorBefore: true }),
      ],
    },
    {
      label:   'Edit',
      entries: [
        entry('undo', 'Undo', ctrl('Z')),
        entry('redo', 'Redo', ctrl('Y'), { enabled: false }),
        entry('cut', 'Cut', ctrl('X'), { separatorBefore: true }),
        entry('copy', 'Copy', ctrl('C')),
        entry('paste', 'Paste', ctrl('V')),
        entry('delete', 'Delete', { key: 'Del' }),
        entry('move-up', 'Move up', { shift: true, key: 'Up' }, { separatorBefore: true }),
        entry('move-down', 'Move down', { shift: true, key: 'Down' }),
      ],
    },
    {
      label:   'Run',
      entries: [
        entry('load', 'Load playlists', ctrl('R')),
        entry('pause', 'Stop downloads', ctrl('P'), { separatorBefore: true }),
        entry('stop', 'Stop', ctrl('.')),
        entry('retry', 'Retry failed', ctrl('T')),
        entry('add-source', 'Add source…', ctrl('L'), { separatorBefore: true }),
        entry('accept-all', 'Accept all songs', ctrl('D')),
        entry('open-folder', 'Open folder', ctrl('O', true), { separatorBefore: true }),
      ],
    },
    {
      label:   'Help',
      entries: [
        entry('shortcuts', 'Shortcuts', { key: 'F1' }),
        entry('docs', 'Documentation'),
        entry('report', 'Report a problem…'),
        entry('about', 'About MVD', undefined, { separatorBefore: true }),
      ],
    },
  ],
  toolbar: [
    { id: 'pause', label: '❚❚ Stop downloads', role: 'other', enabled: true },
    { id: 'stop', label: '■ Stop', role: 'other', enabled: true },
    { id: 'retry', label: 'Retry failed', role: 'other', enabled: false, disabledReason: 'Nothing has failed.' },
    { id: 'add-source', label: '+ Add source', role: 'default', enabled: true },
  ],
  sources: {
    heading:  '3 sources',
    selected: 1,
    rows:     [
      { id: 's1', label: '90s UK Dance Hits', detail: '48 songs', tone: 'good' },
      { id: 's2', label: 'Daft Punk – Discovery', detail: '14 songs', tone: 'busy' },
      { id: 's3', label: 'open.spotify.com/playlist/37i9…', detail: 'could not be read', tone: 'bad' },
    ],
  },
  songs: {
    heading:  '14 songs',
    selected: 1,
    rows:     [
      { id: 'a', ticked: true, artist: 'Daft Punk', title: 'One More Time', progress: 1, status: 'official', tone: 'good' },
      { id: 'b', ticked: true, artist: 'Daft Punk', title: 'Aerodynamic', progress: 0.34, status: 'downloading', tone: 'busy' },
      { id: 'c', ticked: true, artist: 'Daft Punk', title: 'Digital Love', progress: 'unknown', status: 'finding', tone: 'busy' },
      { id: 'd', ticked: true, artist: 'Daft Punk', title: 'Harder, Better, Faster, Stronger', progress: null, status: 'waiting', tone: 'dim' },
      { id: 'e', ticked: false, artist: 'Daft Punk', title: 'Crescendolls', progress: null, status: 'skipped', tone: 'dim' },
      { id: 'f', ticked: true, artist: 'Daft Punk', title: 'Nightvision', progress: null, status: 'failed', tone: 'bad' },
      { id: 'g', ticked: true, artist: 'Daft Punk', title: 'Superheroes', progress: 1, status: 'have it', tone: 'good' },
      { id: 'h', ticked: true, artist: 'Daft Punk', title: 'High Life', progress: null, status: 'waiting', tone: 'dim' },
      { id: 'i', ticked: true, artist: 'Daft Punk', title: 'Something About Us', progress: null, status: 'waiting', tone: 'dim' },
      { id: 'j', ticked: true, artist: 'Daft Punk', title: 'Voyager', progress: null, status: 'waiting', tone: 'dim' },
      { id: 'k', ticked: true, artist: 'Daft Punk', title: 'Veridis Quo', progress: null, status: 'waiting', tone: 'dim' },
      { id: 'l', ticked: true, artist: 'Daft Punk', title: 'Short Circuit', progress: null, status: 'waiting', tone: 'dim' },
      { id: 'm', ticked: true, artist: 'Daft Punk', title: 'Face to Face', progress: null, status: 'waiting', tone: 'dim' },
      { id: 'n', ticked: true, artist: 'Daft Punk', title: 'Too Long', progress: null, status: 'waiting', tone: 'dim' },
    ],
  },
  details: {
    heading: 'Details',
    lines:   [
      'Aerodynamic – Daft Punk (Discovery, 2001)',
      'Official video: youtube.com/watch?v=L93-7bjXKcs',
      'Downloading: 34% of 41.2 MB at 3.1 MB/s, 9 s left',
      '[official] found the link in the video description',
      '[ytdlp] format 137+140 (1080p, m4a)',
    ],
  },
  status: {
    summary: 'Downloading 3 of 14 · 2 done · 1 failed',
    hints:   [
      { shortcut: { key: 'F1' }, label: 'shortcuts' },
      { shortcut: ctrl('L'), label: 'add source' },
      { shortcut: ctrl('P'), label: 'stop downloads' },
    ],
  },
  busy: false,
}
