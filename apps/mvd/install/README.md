# install

Puts MVD where the system keeps programs, and keeps that copy current: the offer to move
the app, the update of an installed copy, the shortcut or bundle, and the entry in the
system's list of programs.

One outcome, one slice. Files by role:

| Role | Files |
|---|---|
| Decisions | `install_location_policy.go` (where it can go, whether to offer), `install_marker_policy.go` (what was asked, which version) |
| Flow | `install_location_use_case.go` (the offer), `update_installed_use_case.go` (replace an older install), `moved_program_cleanup_use_case.go` (remove the old copy once it has exited) |
| Conversions | `install_footprint_mapper.go`, `launcher_entries_mapper.go`, `shortcut_command_mapper.go`, `uninstall_entry_mapper.go` |
| System | `install_location_client.go`, `elevation_state_client_*.go`, `uninstall_entry_client_*.go` |

## Exception to the 12-file review (CLAUDE.md, size)

- **Path:** `apps/mvd/install` (14 production files).
- **Rule waived:** a slice is reviewed at 12 production files.
- **Why:** it is one outcome. Two of its clients exist twice, as `_windows.go` and `_other.go`
  (the build tags Go mandates), so the cohesive count is 12. Splitting the flow would make
  the offer and the update import each other (`updateOffer` takes the move environment and
  `moveOffer` calls `updateOffer`), which is the cycle the rules forbid.
- **Temporary:** no. Revisit if a second outcome appears (for instance uninstalling moving
  here from `apps/mvd/uninstall`).
