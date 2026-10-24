package main

import (
	"errors"
	"fmt"
	"os"
)

type pemFile struct {
	path string
	data []byte
	mode os.FileMode
}

func writeCertAndKeyPEM(certPath, keyPath string, certPEM, keyPEM []byte) error {
	if certPath == keyPath {
		data := make([]byte, 0, len(certPEM)+len(keyPEM))
		data = append(data, certPEM...)
		data = append(data, keyPEM...)
		return writePEMFiles(pemFile{certPath, data, 0600})
	}
	return writePEMFiles(
		pemFile{certPath, certPEM, 0644},
		pemFile{keyPath, keyPEM, 0600},
	)
}

// writePEMFiles never opens an existing target for writing. On a handled error,
// it rolls back only files created by this call. This is not a crash-atomic
// transaction across multiple files.
func writePEMFiles(files ...pemFile) (err error) {
	type createdFile struct {
		path string
		info os.FileInfo
	}
	var created []createdFile
	defer func() {
		if err == nil {
			return
		}
		for i := len(created) - 1; i >= 0; i-- {
			file := created[i]
			info, statErr := os.Lstat(file.path)
			if os.IsNotExist(statErr) {
				continue
			}
			if statErr != nil {
				err = errors.Join(err, fmt.Errorf("cannot check incomplete PEM file %q for cleanup: %w", file.path, statErr))
				continue
			}
			if !os.SameFile(file.info, info) {
				err = errors.Join(err, fmt.Errorf("PEM target %q changed; refusing to remove it during cleanup", file.path))
				continue
			}
			if removeErr := os.Remove(file.path); removeErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to remove incomplete PEM file %q: %w", file.path, removeErr))
			}
		}
	}()

	for _, file := range files {
		f, openErr := os.OpenFile(file.path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, file.mode)
		if openErr != nil {
			if os.IsExist(openErr) {
				return fmt.Errorf("refusing to overwrite existing PEM target %q: %w", file.path, openErr)
			}
			return fmt.Errorf("failed to create PEM file %q: %w", file.path, openErr)
		}
		info, statErr := f.Stat()
		if statErr != nil {
			closeErr := f.Close()
			return fmt.Errorf("cannot identify newly created PEM file %q; file left in place: %w", file.path, errors.Join(statErr, closeErr))
		}
		created = append(created, createdFile{file.path, info})
		_, writeErr := f.Write(file.data)
		closeErr := f.Close()
		if writeErr != nil || closeErr != nil {
			return fmt.Errorf("failed to save PEM file %q: %w", file.path, errors.Join(writeErr, closeErr))
		}
	}
	return nil
}
