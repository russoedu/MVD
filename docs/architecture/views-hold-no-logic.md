# Views hold no logic

Issue #344. The screens of MVD hold no behaviour. What a screen does lives in a renderer-neutral
package; a renderer (the terminal today, a web page with the Windows 95, Mac OS 9 and Windows 3.1
looks later, #345) only turns input into calls and the state into pixels or cells.

## The layers

```
engine, runstate                  the core: runs, events, plain state             (already neutral)
<screen>/ (renderer-neutral)      one screen's behaviour: state, intents, effects  (this work)
terminal adapter (Bubble Tea)     keys and mouse in, cells out; owns the widgets and the cell layout
web adapter (later, #345)         JSON out and in; React draws it
```

## What a neutral package is

1. Its package comment contains the phrase **renderer-neutral**. That marks it: the test in
   `libs/mvd-core/architecture` finds every marked package and fails if any of its dependencies,
   direct or transitive, is a terminal or drawing package (Bubble Tea, Lip Gloss, Bubbles, `look`,
   `cells`, `screenview`, `keybar`, `pointer`, `dialog`, `menubar`, `lineedit`, `tui`, `internal/...`).
2. **State** is a struct (`Session`, or `Screen` for a small one) with exported fields for what a
   renderer may read. It is a value: a copy of a screen keeps its own state, so a map or slice is
   replaced, never changed in place. The terminal model holds it by value.
3. **Intents** are methods named for what the person means: `TickSong(id, extend)`, `SortBy(column)`,
   `AcceptSongs()`, `ToggleFollow()`. Never a key, a coordinate, a cell or a pixel. A method changes
   the state (pointer receiver) and, when the host must do something, returns an **Effect**.
4. **Effects** are plain values (`OpenURL{URL}`, `Quit{}`, ...), in an `effect_contract.go`. The
   terminal adapter turns them into `tea.Cmd`s; the web adapter into events. A neutral package
   never returns a `tea.Cmd`.
5. **Text being edited** is a string in the state. The widget (a Bubbles text area, a DOM textarea)
   belongs to the renderer and pushes its value up with an intent such as `SetLinksText(s)`.
6. **Words the person reads** that every renderer should say alike (a notice such as "3 songs
   accepted: they will download", a status word) are in the neutral package. How they look is not.
7. Things that need the engine take an interface the package declares (`Controller`, `Chooser`,
   `SourceAdder`); the host passes the real run.

## What stays in the renderer

The test: *would a second renderer have to agree on it?* If yes it is in the neutral package, if
no it stays in the adapter.

- Sizes in cells or pixels: column and pane widths, the dragging of a border, scroll offsets, the
  height of a pane, the marquee, the spinner frame.
- Hit-testing by position, hover, pressed, focus rings, and whether a popup is only a visual
  overlay.
- Colours and glyphs (a tone or a kind is neutral; the colour or the glyph is the renderer's).
- The key that does an intent (`d` accepts in the terminal; a button or a menu in the web).
- The widgets themselves.

## How to extract one area

1. Read the area; list its state, its decisions, its effects and its drawing.
2. Create the neutral package beside it with the files named by role (`*_model.go`,
   `*_use_case.go`, `*_policy.go`, `*_contract.go`, `*_algorithm.go`), at most 12 production files.
3. Move state and decisions in; leave the adapter translating `tea.Msg` to intents and the state to
   cells. Do it in steps that keep `go test ./...` green.
4. Move the behaviour tests into the neutral package, driving intents. The terminal tests that
   remain are key and mouse mapping, and the golden drawings.
5. The golden DOS drawings and the screen tests must pass **without being edited**, except for
   renamed fields.
6. `refactor:` commits (they do not release), `Part of #344` in the PR, one area per PR.

## Areas

| Area | Where it is today |
|---|---|
| song list rules | `songlist` (done, #346) |
| download screen | `run/download` |
| setup screens | `setup/*` |
| playlist editor | `editor/*` |
| preferences popup, menu bar, dialogs | `prefspopup`, `menubar`, `dialog`, `lineedit` |
| workspace and app flow | `workspace`, `tui`, `apps/mvd/terminalui` |
