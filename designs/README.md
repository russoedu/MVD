# Designs

Every screen of MVD is a [Cuppa](https://github.com/meta-tui/cuppa) design (`*.cuppa`). Cuppa exports them as the Go package `libs/mvd-core/screens`, which the screens of `libs/mvd-core/tui` fill in and print.

- Open a design with `cuppa-tui designs/advanced.cuppa` (or the desktop or web app), change it, save it.
- Export every design again: `cuppa screens designs -o libs/mvd-core/screens -p screens`.
- Check that the generated files are what the designs export (nothing is written; exit 1 when they differ): add `--check` to the same command. It is not in CI: `ci.yml` is generated and `cuppa` is not published as a binary, so run it before committing a design.
- Check that nothing changed in how a screen looks: `go test ./libs/mvd-core/tui -run LegacyScreens`.
- Commit the designs and the generated files together, naming the `cuppa-tui` version that wrote them.

Inputs that only a layout expression or a condition reads (`LeftWidth`, `FooterLines`, `HelpWidth`, ...) have a sample value in the design (`samples`, cuppa-tui 0.0.66+; the designer's "Layout inputs" panel), so a preview opens at a realistic size. The app sets every one of them each frame, so the sample is only what the designer shows and the default of `Default<Screen>Props`: a model that reads that default must set the input itself (`FooterLines` is reset in `downloadProps`).

`mvd.cuppa` is the 11 designs as one multi-page design (a page per screen, canvas 120x36), made to try Cuppa's slice layout (issue #180): `cuppa-tui designs/mvd.cuppa`, then the pages bar under the canvas. The 11 single designs stay the source of the screens until the slice export is adopted; it is not exported yet.

Generated files are never edited by hand. See `nerdy-info.md`, section 2.1.
