/*
 * The classic themes: which exist, the layout rules of each chrome family, and each theme's
 * tokens. A theme file sets custom properties on :root[data-theme=...] (fonts, colours, bevels,
 * sizes, animations) and styles the scroll bars; the window's components read only those
 * properties, and put what differs in structure under [data-family=...].
 */
import './classic-fonts.style.css'
import './windows-3-1.style.css'
import './windows-95.style.css'
import './windows-xp.style.css'
import './windows-vista.style.css'
import './mac-os-7.style.css'
import './mac-os-9.style.css'
import './aqua.style.css'
import './classic-controls.style.css'

export { classicThemeOf, classicThemes, type ChromeFamily, type ClassicTheme, type ClassicThemeId } from './classic-theme.model'
export { layoutOf, orderDialogButtons, type ChromeLayout, type WindowControl } from './chrome-layout.policy'
export { shortcutLabel } from './shortcut-label.mapper'
export { isMnemonic, pressedLetter, splitMnemonic, type MnemonicParts } from './mnemonic.mapper'
export { pixelSnapZoom } from './pixel-snap.algorithm'
