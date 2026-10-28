package selfupdate

import (
	"archive/zip"
	"errors"
	"io"
	"os"
)

// maxProgramBytes is the most the program inside a zip may be once unpacked.
const maxProgramBytes = 300 << 20

// extractProgram writes the program found in the zip at zipPath to dest. The zip is only read for
// that one name, never for paths of its own, so it cannot write anywhere else.
func extractProgram(zipPath, name, dest string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	for _, f := range r.File {
		if f.Name != name {
			continue
		}
		in, err := f.Open()
		if err != nil {
			return err
		}
		defer func() { _ = in.Close() }()
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o700)
		if err != nil {
			return err
		}
		n, err := io.Copy(out, io.LimitReader(in, maxProgramBytes+1))
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
		if err == nil && n > maxProgramBytes {
			err = errors.New("the program in the download is larger than any MVD release")
		}
		if err != nil {
			_ = os.Remove(dest)
		}

		return err
	}

	return errors.New(name + " is not in the download")
}
