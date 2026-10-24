//go:build !windows

package main

import (
	"errors"
	"os"
)

func createPrivatePFXFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
}

func verifyPFXPermissions(f *os.File) error {
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.Mode().Perm()&0077 != 0 {
		return errors.New(tr("PFX permissions allow access by other users"))
	}
	return nil
}
