# Classic interface themes: plan

Status: plan (2026-10-26). Tracked in #345. Depends on #344, which takes the logic out of the views.

MVD gets seven looks styled after classic desktops, beside the DOS look it has today:

| Theme | Era | Chrome family |
|---|---|---|
| Windows 3.1 | 1992 | Windows classic |
| Windows 95 | 1995 | Windows classic |
| Windows XP (Luna) | 2001 | Windows themed |
| Windows Vista (Aero) | 2007 | Windows themed |
| Mac OS 7.0 | 1991 | Mac classic |
| Mac OS 9.0 (Platinum) | 1999 | Mac classic |
| Mac OS X 10.0 (Aqua) | 2001 | Aqua |

It is a simulation inside MVD's own window. Where the real OS put something outside the window, such as the Mac menu bar at the top of the screen or the Windows taskbar, we keep it inside the window and do not try to copy the rest of the desktop.

## Principles

1. **Behaviour stays in Go.** The screens' state, rules and intents are the renderer-neutral packages from #344. A theme only draws a snapshot and turns clicks and keys into intents. No theme decides anything about songs, sources or runs.
2. **A theme is tokens plus chrome.** CSS custom properties hold colours, metrics and fonts. Each *chrome family* has its own React components for the window frame, title bar, menu bar, scroll bar, buttons and dialogs, because these differ in structure and not only in colour. The seven themes need four families. Two themes share a family when they differ mainly in tokens and a few parts.
3. **React Aria Components** give the accessible, unstyled widgets: menu bar, table, list box, dialog, radio group, text field, progress bar, check box. They carry no look of their own, so every pixel comes from the theme.
4. **DOS stays.** It is still the Bubble Tea view through TReactUI, and it can be picked like any other theme.
5. **No copied artwork.** We draw our own icons and glyphs in each period's style. We do not ship Microsoft or Apple bitmaps, the Apple logo or the Windows flag.

## What the snapshot must carry

So that themes can place things the way their OS did, the JSON snapshot describes *roles*, not positions:

- **Buttons carry a role** (`default`, `cancel`, `other`) and an enabled flag with the reason when disabled. Windows themes place them `[OK] [Cancel]`. Mac themes place them `[Cancel] [OK]`, with the default button rightmost.
- **Menu entries carry their shortcut as data** (`{ctrl: true, key: "L"}`), not as text. Windows themes write `Ctrl+L` aligned right. Mac themes write `⌘L`.
- **Mnemonics.** A menu entry carries its access letter, so Windows themes can underline it and handle Alt+letter. Mac themes ignore it.
- **Busy state.** The snapshot says what MVD is doing (`activity`: idle, working in the background, or keeping the user waiting), so each theme can show its own busy pointer: hourglass, watch, spinning wait cursor.
- **Progress is a number**, or "unknown", which gets the theme's indeterminate bar.

## Behaviours that differ

All of these are renderer-side conventions: how a click or key is read and how something looks while it happens. None is a business rule.

### Window controls

| Theme | Controls | Title | Double-click on title bar |
|---|---|---|---|
| Win 3.1 | Control-menu box on the left (bar glyph). Minimise ▼ and maximise ▲ on the right. No close button: close is in the control menu, or a double-click on the control-menu box. | Centred | Maximise / restore |
| Win 95 | App icon on the left (opens the system menu). `_ □ ×` on the right, with a gap before ×. | Left-aligned, bold | Maximise / restore |
| XP | As 95, in rounded blue Luna buttons. Close is red. | Left, bold, with a shadow | Maximise / restore |
| Vista | As XP. Glass caption buttons. Close is wider, red and glowing. | Left | Maximise / restore |
| Mac 7.0 | Close box on the left. Zoom box on the right. No collapse box: WindowShade came with 7.5. | Centred, on horizontal stripes | Nothing |
| Mac 9 | Close box on the left. Zoom and collapse boxes on the right. | Centred, on stripes | Collapse (window shade) |
| Aqua 10.0 | Red, yellow and green traffic lights on the left: close, minimise, zoom. | Centred | Minimise |

The controls act on MVD's real window through Wails: minimise, maximise or zoom, close.

**Mac 9 collapse:** the simplest version minimises the window. A faithful version shrinks the OS window to its title bar. To be decided when we build it.

### Menus

| Theme | How a menu opens | Look |
|---|---|---|
| Win 3.1 / 95 | Click opens it and it stays open (sticky). Alt or F10 focuses the menu bar. Alt+letter opens a menu. | Mnemonic underlined. Shortcuts aligned right. Highlight is navy with white text. |
| XP / Vista | Sticky, as 95. Mnemonics are hidden until Alt is pressed. | XP: flat menu with a blue highlight. Vista: soft gradient highlight with a rounded edge. |
| Mac 7.0 | **Press and hold.** The menu is open only while the button is held, and releasing over an entry chooses it. | Inverted highlight (black with white text). `⌘` shortcuts. |
| Mac 9 | Sticky click, as Mac OS 8 introduced. | Platinum, with a lavender highlight. |
| Aqua | Sticky. | Translucent white menu with a blue gel highlight and a drop shadow. |

The menu bar sits at the top of MVD's window in every theme, including the Mac ones.

### Buttons, focus and the keyboard

- **Default button.** Win 3.1/95: thicker black border. XP: blue inner glow. Vista: blue glow that pulses slowly. Mac 7/9: thick rounded outline ring around the button. Aqua: blue gel that **pulses** about once a second.
- **Keyboard focus.** Windows themes draw a dotted focus rectangle on the focused control, and Tab moves between all controls. Mac 7/9 had no focus on buttons: Tab moved only between text fields, Return pressed the default button and Esc or ⌘. pressed Cancel. Aqua 10.0 works like Mac 9 by default.
- **We keep full keyboard access in every theme** for accessibility. In Mac themes, the focus ring is drawn only when the keyboard is being used (`:focus-visible`), so they still look right with the mouse.
- **Shortcuts on Mac themes.** They show `⌘`. On Windows and Linux the key is still Ctrl. On a real Mac both Cmd and Ctrl work.

### Scroll bars

| Theme | Arrows | Thumb |
|---|---|---|
| Win 3.1 | One at each end | Fixed-size square (not proportional) |
| Win 95 / XP / Vista | One at each end | Proportional |
| Mac 7.0 | One at each end | Fixed-size square, grey track |
| Mac 9 | One at each end | Proportional (the Appearance "smart scrolling" option) |
| Aqua 10.0 | Both together at the bottom or right end | Proportional blue gel "pill" |

Clicking the track pages, and dragging the thumb scrolls, in every theme. MVD's Go scroll bar already behaves this way. The themes only draw it differently.

### Selection, lists and tables

- **Win 3.1 / 95:** navy highlight with white text. Column headers are raised buttons.
- **XP:** `#316AC5` highlight. Headers are flat, with an orange bar under the column the pointer is over.
- **Vista:** light blue gradient, with a rounded hover and selection rectangle.
- **Mac 7:** inverted (black with white text).
- **Mac 9:** lavender highlight, Platinum headers.
- **Aqua:** blue highlight. Rows alternate white and pale blue.

### Progress bars

- **Win 3.1:** solid fill with the percentage written in the middle, inverted where the fill passes under the text.
- **Win 95:** navy blocks.
- **XP:** green blocks in a rounded trough.
- **Vista:** a continuous green bar with a light sweep that keeps moving.
- **Mac 7:** solid black fill in a white trough.
- **Mac 9:** blue-grey fill. Unknown progress is a moving striped "barber pole".
- **Aqua:** blue gel with stripes that keep moving. Unknown progress is a striped pill.

### Motion

- **Win 3.1, Win 95, Mac 7, Mac 9:** nothing animates. Menus and dialogs appear at once.
- **XP:** menus fade in. Buttons light up on hover.
- **Vista:** buttons fade in their glow over about 150 ms. Progress bars sweep.
- **Aqua:** the default button pulses. Dialogs come down as **sheets**, sliding out from under the title bar and attached to the window instead of floating in the middle. Progress bars animate. The genie effect is out of scope.

Every animation is turned off under `prefers-reduced-motion`.

### Dialogs

- **Windows themes:** dialogs are centred, with their own title bar and close button. The buttons sit bottom-right (95 onwards) or centred (3.1).
- **Mac 7/9:** a modal dialog has **no title bar**. It is a double-bordered box in the top third of the screen, with the default button bottom-right.
- **Aqua:** a sheet that belongs to the window.

The "Add music" window and the other questions follow their theme's rule. The texts and buttons come from the same snapshot in every theme.

## CSS techniques

- **Tokens.** Each theme is one set of custom properties on `:root[data-theme="win95"]` and the other theme names: colours, border widths, title and menu heights, scroll bar width, fonts. The family components read only tokens.
- **Bevels (Windows classic, Mac 9).** Two-pixel 3D edges come from layered `box-shadow: inset` lines, light at the top-left and dark at the bottom-right, swapped when pressed. This is the 98.css technique, MIT licensed and used here as a reference for metrics.
- **Title-bar stripes (Mac 7/9) and pinstripes (Aqua).** `repeating-linear-gradient`. No images.
- **Aqua gel.** Layered gradients: a base colour, a bright reflection on the upper half and a soft glow at the bottom, plus an inner shadow. The pulse is a `@keyframes` animation on the gel's brightness.
- **Vista glass.** A translucent frame colour with `backdrop-filter: blur()`. It falls back to a solid tint where blur is not supported.
- **Pixel-exact rendering for the bitmap-era themes (3.1, 95, Mac 7, Mac 9).** Fonts at their native pixel size, `-webkit-font-smoothing: none`, `font-smooth: never`, and only integer scaling: the whole UI scales by 1×, 2× or 3× with `zoom` and `image-rendering: pixelated` on icons. Fractional scaling makes pixel fonts blur, so these themes do not use it.
- **Icons.** Our own 16×16 and 32×32 pixel art for each family, as SVG drawn on the pixel grid with `shape-rendering: crispEdges`.
- **Cursors.** Custom `cursor: url(...)` for the busy state in each family: hourglass, watch, spinning wait cursor.
- **Reference libraries for metrics, not dependencies:** 98.css (Win 98), XP.css, 7.css (Vista/7), all MIT, and system.css (Mac System 6). Their class-based CSS is not mixed into our components.

## Fonts

MVD runs on the user's own computer, so each font stack tries the installed system font first with `local()`, then a bundled open font, then a generic fallback. Windows already ships Tahoma, Trebuchet MS, Segoe UI, Lucida Sans Unicode, Lucida Console and Courier New. macOS ships Lucida Grande, Monaco and Geneva. We bundle only fonts whose licence allows embedding.

| Theme | UI text (original) | Titles and menus (original) | Log pane (original) | Bundled open fallback |
|---|---|---|---|---|
| Win 3.1 | MS Sans Serif 8 | System (bold) | Fixedsys | W95FA, OFL 1.1. No open recreation of the bold System font is confirmed yet. Fixedsys Excelsior, public domain, for the log. |
| Win 95 | MS Sans Serif 8 | MS Sans Serif 8 bold | Courier New / Fixedsys | **W95FA** (OFL 1.1, also on Fontsource as `win95fa`). Fixedsys Excelsior. |
| XP | Tahoma 8 | Trebuchet MS 10 bold | Lucida Console | Installed fonts on Windows. **Wine Tahoma** (LGPL) elsewhere: check whether LGPL allows embedding it as a web font. |
| Vista | Segoe UI 9 | Segoe UI 9 | Consolas | Installed on Windows. **Selawik** (Microsoft's open Segoe UI fallback, OFL 1.1) elsewhere. |
| Mac 7.0 | Geneva 9 | Chicago 12 | Monaco 9 | **ChicagoFLF**, public domain, a vector Chicago. Sysfont (Alina Sava) and Chicago Kare are closer to the bitmap: licences to check. Geneva: no open recreation confirmed yet. Monaco is installed on macOS. |
| Mac 9 | Geneva 10 | Charcoal 12 | Monaco 9 | **Virtue**, a free Charcoal look-alike: licence to check. Geneva and Monaco as for Mac 7. |
| Aqua 10.0 | Lucida Grande 13 / 11 | Lucida Grande 13 | Monaco 10 | Lucida Grande is installed on macOS. On Windows, **Lucida Sans Unicode** is installed and is the design Lucida Grande grew from, so the stack is `local("Lucida Grande"), local("Lucida Sans Unicode")`. |

Before any font is bundled, its licence file goes into the repo beside it and is listed in the About window's credits.

## Colours (starting values)

These are approximate. Step 2 replaces them with values sampled from reference screenshots.

| Theme | Face | Window | Active title | Selection | Accent |
|---|---|---|---|---|---|
| Win 3.1 | `#C0C0C0` | `#FFFFFF` | `#000080`, white text | `#000080` | — |
| Win 95 | `#C0C0C0`. Bevel `#FFFFFF` / `#808080` / `#000000`. | `#FFFFFF` | `#000080` | `#000080` | Desktop `#008080` |
| XP | `#ECE9D8` | `#FFFFFF` | Blue gradient `#0058E6` → `#3A93FF` | `#316AC5` | Close `#E04343`. Progress `#2FB52F`. |
| Vista | `#F0F0F0` | `#FFFFFF` | Glass, tinted and blurred | `#CCE8FF`, border `#99D1FF` | Progress `#06B025` |
| Mac 7.0 | `#FFFFFF` (`#DDDDDD` in colour mode) | `#FFFFFF` | Black stripes on white | Inverted | — |
| Mac 9 | Platinum `#DDDDDD` | `#FFFFFF` | Grey stripes | Lavender `#CCCCFF` | Progress `#6666CC` |
| Aqua 10.0 | Pinstripe `#ECECEC` / `#F7F7F7` | `#FFFFFF` | Pinstripes | `#3875D7` | Gel blue `#3E8FE4`. Lights red, yellow, green. |

## What every theme draws

Everything here is in MVD today: the menu bar, the toolbar buttons (Start, Pause, Stop, + Add source and the rest), the sources pane, the songs table with check boxes, progress and status, the details and technical log pane with its divider, the scroll bars, the status and key-hint line, the "Add music" window with its radio buttons and text fields, the preferences, the About window, and the one-line questions such as quit and stop.

## Steps

1. **Finish #344** (the workspace run lifecycle, `tui/app_flow_model.go`, `apps/mvd/terminalui`), so every screen is a snapshot plus intents.
2. **Reference board.** For each theme, screenshots of its real windows, menus, buttons, scroll bars, progress bars and dialogs. Exact metrics (title height, border widths, scroll bar width, menu height) and sampled colours go into the token files.
3. **JSON channel** between Go and the page (Wails events): the snapshot goes out, intents come in. The snapshot gains the roles listed in "What the snapshot must carry".
4. **React renderer** on React Aria Components, with the theme switch and the token loader.
5. **Windows classic family:** Windows 95 first, then Windows 3.1 as a variant (other control boxes, fixed scroll thumb, bold System font, no taskbar look).
6. **Mac classic family:** Mac OS 9 first, then Mac OS 7.0 as a variant (no collapse box, press-and-hold menus, inverted selection, black and white).
7. **Windows themed family:** XP, then Vista (glass, glow, fades).
8. **Aqua:** Mac OS X 10.0 (gel, pinstripes, pulsing default button, sheets, translucent menus).
9. **Theme picker** in the preferences, with a small preview of each theme. DOS stays selectable.

Each family is one PR or a small series. A theme is done when MVD's screens are drawn in it, every intent works by mouse and keyboard, and a screenshot test pins its look.

## Built so far (#369)

The first three themes have fonts that are free to use: Windows 95 (W95FA), Windows Vista (Segoe UI, or Selawik where it is missing) and Mac OS X 10.0 (the installed Lucida Grande, or Lucida Sans Unicode). They are built in `apps/mvd-web` and drawn from a sample snapshot at `/themes`, until the program sends its own (steps 1 and 3).

| Slice | Holds |
|---|---|
| `workspace-snapshot/` | The contract: what the window shows (roles, not places) and the intents it sends back. |
| `classic-themes/` | The theme list, each family's layout rules (`chrome-layout.policy.ts`), how shortcuts are written, the bundled fonts, one token file per theme and the shared controls. |
| `classic-window/` | Frame, title bar, menu bar, toolbar and status line. |
| `classic-panes/` | Sources list, songs table with progress bars, details pane. |
| `add-music-window/` | The Add music window: a dialog on Windows, a sheet on a Mac. |
| `theme-preview/` | The `/themes` page and its sample list. |

How the CSS is split: a theme file sets custom properties on `:root[data-theme=...]` and styles that theme's scroll bars. Component stylesheets read only those properties, and put what differs in structure under `[data-family=...]`. The theme is set on the page root, so menus, which open outside the window, wear it too.

Added in #371:
- **Keyboard letters.** Menus, menu entries, dialog buttons, choices and fields can carry a `mnemonic`. Windows 95 always underlines it. Vista underlines it once Alt or F10 is pressed and hides it again on a click (the layout's `mnemonics`: `always`, `on-alt` or `none`). On Windows:
  - Alt and a menu's letter opens that menu with its first entry lit.
  - A letter in an open menu picks that entry, and Left or Right moves to the next menu.
  - Alt or F10 alone puts the focus on the menu bar.
  - In the dialog, Alt and a letter picks a choice, focuses a field or presses a button. The letter alone does the same when the focus is not in a text field.
  - A Mac has none of this.
- **Busy pointers.** The snapshot's `activity` is `idle`, `background` or `waiting`.
  - Each theme draws its own pointers as SVG in `classic-themes/cursors/`, made for MVD: the Windows 95 hourglass and arrow-with-hourglass, the Vista ring and arrow-with-ring, and the Mac OS X spinning wheel. A Mac shows no pointer for work in the background.
  - Browsers draw image pointers still, so the ring and the wheel do not turn.
- **Sharp Windows 95.** The family's `pixelSnap` zooms the page so one CSS pixel is a whole number of screen pixels: 80 % on a screen at 125 %, 133 % at 150 %. It follows changes of scale.
  - React Aria places menus by screen position, which the zoom would scale again. The menu's box therefore undoes the zoom (`--pixel-zoom`), and the panel inside puts it back.

Still to do for these three themes:
- **Aqua scroll bars:** Chrome does not draw both arrows together at the end, so the Aqua bar has no arrows yet.
- **Menu bar tracking:** with a menu open, moving the pointer over another menu title does not open that menu yet.

## Open questions

- **Theme names.** "Windows" and "Mac OS" are trademarks. We could show neutral names in the picker ("Classic 95", "Platinum", "Aqua-style") and keep the OS names in the code and docs.
- **Mac 9 collapse:** minimise the window, or shrink it to its title bar?
- **Fonts still to find or clear:** an open Geneva 9 recreation, an open bold Windows 3.1 System font, and the licences of Virtue, Sysfont and Chicago Kare.
- **Sounds** (Chimes, Sosumi): out of scope unless wanted.
