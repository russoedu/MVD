package programfile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Place copies the program at src to dst, creating dst's folder. It writes to a
// temporary name first, so a copy that is interrupted never leaves half a program where
// a shortcut points. An existing program at dst is replaced, which fails, saying so, if
// it is running.
func Place(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	partial := dst + ".part"
	out, err := os.OpenFile(partial, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(partial)

		return err
	}

	if _, statErr := os.Stat(dst); statErr == nil {
		if err := os.Remove(dst); err != nil {
			_ = os.Remove(partial)

			return fmt.Errorf("the copy already there cannot be replaced, probably because it is running: %w", err)
		}
	}
	if err := os.Rename(partial, dst); err != nil {
		_ = os.Remove(partial)

		return err
	}

	return nil
}
