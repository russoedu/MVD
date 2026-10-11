# Designs

Every screen of MVD is a [Cuppa](https://github.com/meta-tui/cuppa) design (`*.cuppa`). Cuppa exports them as the Go package `libs/mvd-core/screens`, which the screens of `libs/mvd-core/tui` fill in and print.

- Open a design with `cuppa-tui designs/advanced.cuppa` (or the desktop or web app), change it, save it.
- Export every design again: `cuppa screens designs -o libs/mvd-core/screens -p screens`.
- Check that nothing changed in how a screen looks: `go test ./libs/mvd-core/tui -run LegacyScreens`.
- Commit the designs and the generated files together, naming the `cuppa-tui` version that wrote them.

Generated files are never edited by hand. See `nerdy-info.md`, section 2.1.
