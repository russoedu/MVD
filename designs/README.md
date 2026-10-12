# Designs

`mvd.cuppa` is a picture of the screens of MVD: one page per screen, drawn in [Cuppa](https://github.com/meta-ui-labs/cuppa), a design tool. The build does not read it, and nothing in the code depends on Cuppa.

The screens themselves are Go, maintained by hand: a package per page in its feature folder (`libs/mvd-core/setup`, `run`, `editor`), a `<page>_page.go` with the layout, the shared drawing code once in `libs/mvd-core/internal`, and the models beside each page. They were first written by Cuppa's Go export and are no longer regenerated; to change how a screen looks, edit its Go file and check it with `go test ./libs/mvd-core/...` (each slice keeps reference drawings in its `testdata/`, compared cell by cell by `libs/mvd-core/testing/drawing`).

If you change a screen's layout and want the picture to match, open `mvd.cuppa` in Cuppa and update it by hand. See `nerdy-info.md`, section 2.1.
