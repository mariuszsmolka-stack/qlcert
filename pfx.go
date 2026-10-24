package main

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"golang.org/x/term"
	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

func promptPFXPassword(input *os.File, output io.Writer) ([]byte, error) {
	fd := int(input.Fd())
	if !term.IsTerminal(fd) {
		return nil, errors.New(tr("PFX requires an interactive terminal for hidden password input; no CA was created"))
	}
	return readPFXPassword(output, func() ([]byte, error) { return term.ReadPassword(fd) })
}

// The reader must disable echo. Never accept passwords from flags, env or pipes.
func readPFXPassword(output io.Writer, readHidden func() ([]byte, error)) ([]byte, error) {
	fmt.Fprint(output, tr("PFX password (use a long, unique random password; input is hidden): "))
	password, err := readHidden()
	fmt.Fprintln(output)
	if err != nil {
		clear(password)
		return nil, errors.New(tr("could not read the PFX password without echo"))
	}
	if err := validatePFXPassword(password); err != nil {
		clear(password)
		return nil, err
	}
	fmt.Fprint(output, tr("Confirm PFX password (input is hidden): "))
	confirmation, err := readHidden()
	fmt.Fprintln(output)
	defer clear(confirmation)
	if err != nil || !bytes.Equal(password, confirmation) {
		clear(password)
		return nil, errors.New(tr("PFX password confirmation failed"))
	}
	return password, nil
}

func validatePFXPassword(password []byte) error {
	if !utf8.Valid(password) || len(bytes.TrimSpace(password)) == 0 || bytes.IndexByte(password, 0) >= 0 {
		return errors.New(tr("PFX password must be nonempty valid text without NUL characters"))
	}
	for _, r := range string(password) {
		if r > 0xffff {
			return errors.New(tr("PFX password contains characters unsupported by PKCS#12 (such as emoji)"))
		}
	}
	return nil
}

func encodePFX(key crypto.PrivateKey, certificate, ca *x509.Certificate, password []byte) ([]byte, error) {
	if err := validatePFXPassword(password); err != nil {
		return nil, err
	}
	// Only the leaf key is exported. The CA contributes its public certificate.
	return pkcs12.Modern2023.Encode(key, certificate, []*x509.Certificate{ca}, string(password))
}

func checkPFXTarget(path string) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf(tr("refusing to overwrite existing PFX target %q"), path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf(tr("cannot check PFX target %q: %w"), path, err)
	}
	return nil
}

// A failing write removes only the file created by this operation.
func writePFXFile(path string, data []byte) error {
	f, err := createPrivatePFXFile(path)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf(tr("refusing to overwrite existing PFX target %q"), path)
		}
		return fmt.Errorf(tr("failed to create PFX file %q: %w"), path, err)
	}
	info, err := f.Stat()
	if err != nil {
		return errors.Join(err, f.Close()) // Identity unknown: leave the empty file.
	}
	err = verifyPFXPermissions(f)
	if err == nil {
		_, err = f.Write(data)
	}
	err = errors.Join(err, f.Close())
	if err == nil {
		return nil
	}
	current, statErr := os.Lstat(path)
	if statErr == nil && os.SameFile(info, current) {
		err = errors.Join(err, os.Remove(path))
	} else if !os.IsNotExist(statErr) {
		err = errors.Join(err, statErr, errors.New(tr("PFX target changed or cannot be identified; refusing cleanup")))
	}
	return fmt.Errorf(tr("failed to save private PFX file %q: %w"), path, err)
}
