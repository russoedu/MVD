package runcontrol

// EditFacts are what the workspace knows of the list being edited, for the Edit menu.
type EditFacts struct {
	// WritingAddress says an address is being written in its row; HasSelection that part of it is
	// selected.
	WritingAddress, HasSelection bool
	// PlaylistSelected says the screen has a playlist selected.
	PlaylistSelected bool
	// CanUndo and CanRedo say the history has a step to go back or forward to.
	CanUndo, CanRedo bool
	// Selected is the index of the playlist the cursor is on and HasCursor whether there is one;
	// Count is how many playlists the list has.
	Selected  int
	HasCursor bool
	Count     int
}

// EditMenu says which entries of the Edit menu can be used now.
type EditMenu struct {
	Cut, Copy, Paste, Delete, Undo, Redo, MoveUp, MoveDown bool
}

// EditMenuOf decides the Edit menu. Copy and Paste are for the address being written, and Copy needs
// something selected; the other commands work on the playlist the cursor is on, and wait for the
// downloads to stop.
func EditMenuOf(f Facts, e EditFacts) EditMenu {
	free := f.Editable()
	live := f.Addable()
	writing := (free || live) && e.WritingAddress

	menu := EditMenu{
		Cut:      writing && e.HasSelection,
		Copy:     writing && e.HasSelection,
		Paste:    free || live,
		Delete:   live && !writing && e.PlaylistSelected,
		Undo:     free && !writing && e.CanUndo,
		Redo:     free && !writing && e.CanRedo,
		MoveUp:   free && !writing && e.HasCursor && e.Selected > 0,
		MoveDown: free && !writing && e.HasCursor && e.Selected < e.Count-1,
	}
	if free {
		menu.Delete = !writing && e.HasCursor
	}

	return menu
}
