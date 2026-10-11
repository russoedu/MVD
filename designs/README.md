# Designs

Every screen of MVD is a page of one [Cuppa](https://github.com/meta-ui-labs/cuppa) design, `mvd.cuppa`. Cuppa exports it in the slice layout: a package per page in its feature folder (`libs/mvd-core/setup`, `run`, `editor`) and the shared drawing code once in `libs/mvd-core/internal`. The models of `libs/mvd-core/tui` fill each page's props and print it.

- Open the design with `cuppa-tui designs/mvd.cuppa` (or the desktop or web app), change a page with the pages bar, save it.
- Export again: `cuppa screens designs/mvd.cuppa -o libs/mvd-core --layout slices --library -m youtube-downloader/libs/mvd-core`.
- Check that the generated files are what the designs export (nothing is written; exit 1 when they differ): add `--check` to the same command. It is not in CI: `ci.yml` is generated and `cuppa` is not published as a binary, so run it before committing a design. The export removes only files that start with Cuppa's header.
- Check that nothing changed in how a screen looks: `go test ./libs/mvd-core/tui -run LegacyScreens`.
- Commit the designs and the generated files together, naming the `cuppa-tui` version that wrote them.

The page's folder is set in the designer (pages list, *Folder…*). Inputs that only a layout expression or a condition reads (`LeftWidth`, `FooterLines`, `HelpWidth`, ...) have a sample value per page (the *Layout inputs* panel), so a preview opens at a realistic size; the app sets every one each frame, so a model that reads a page's `Default()` must set the input itself (`FooterLines` is reset in `downloadProps`).

The design has a **master page** (the ◆ entry of the pages list, cuppa-tui 0.0.71+): the title label and the key bar are drawn once there and shown on every page; a page hides what it does not want (the eye in the Layers panel). Their inputs keep their names (`Title`, `Hints`), so each page's `Props` are the same; each page keeps its own text for them (the *Master text* panel of the page, `Page.Values`), which is the default of its `Props`, so a model sets `props.Title` only when the text changes at run time (the list and the editor do). The export writes the master's parts once, in `libs/mvd-core/internal/components/master` (cuppa-tui 0.0.73+), and every page calls them, key bar hit test included.

Generated files are never edited by hand. See `nerdy-info.md`, section 2.1.
