# Designs

Every screen of MVD is a page of one [Cuppa](https://github.com/meta-ui-labs/cuppa) design, `mvd.cuppa`. Cuppa exports it in the slice layout: a package per page in its feature folder (`libs/mvd-core/setup`, `run`, `editor`) and the shared drawing code once in `libs/mvd-core/internal`. The models of `libs/mvd-core/tui` fill each page's props and print it.

- Open the design with `cuppa-tui designs/mvd.cuppa` (or the desktop or web app), change a page with the pages bar, save it.
- Export again: `cuppa screens designs/mvd.cuppa -o libs/mvd-core --layout slices --library -m youtube-downloader/libs/mvd-core`.
- Check that the generated files are what the designs export (nothing is written; exit 1 when they differ): add `--check` to the same command. It is not in CI: `ci.yml` is generated and `cuppa` is not published as a binary, so run it before committing a design. The export removes only files that start with Cuppa's header.
- Check that nothing changed in how a screen looks: `go test ./libs/mvd-core/tui -run LegacyScreens`.
- Commit the designs and the generated files together, naming the `cuppa-tui` version that wrote them.

The page's folder is set in the designer (pages list, *Folder…*). Inputs that only a layout expression or a condition reads (`LeftWidth`, `FooterLines`, `HelpWidth`, ...) have a sample value per page (the *Layout inputs* panel), so a preview opens at a realistic size; the app sets every one each frame, so a model that reads a page's `Default()` must set the input itself (`FooterLines` is reset in `downloadProps`).

Generated files are never edited by hand. See `nerdy-info.md`, section 2.1.
