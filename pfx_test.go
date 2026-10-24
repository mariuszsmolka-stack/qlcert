package main

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

func TestPFXPasswordInput(t *testing.T) {
	secret := "c4a35f072d9a8e192ba1-test-secret"
	for _, tc := range []struct {
		name, password, confirmation string
		failRead                     int
		valid                        bool
	}{
		{name: "matching", password: secret, confirmation: secret, valid: true},
		{name: "unicode_and_spaces", password: "  żółć-unique-password  ", confirmation: "  żółć-unique-password  ", valid: true},
		{name: "empty"}, {name: "whitespace", password: " \t "},
		{name: "mismatch", password: secret, confirmation: "different"},
		{name: "read_error", password: secret, failRead: 1},
		{name: "confirmation_error", password: secret, confirmation: secret, failRead: 2},
		{name: "invalid_utf8", password: "\xff"},
		{name: "nul", password: "a\x00b"},
		{name: "unsupported_unicode", password: "test-🔑"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			reads := 0
			password, err := readPFXPassword(&output, func() ([]byte, error) {
				reads++
				text := tc.password
				if reads == 2 {
					text = tc.confirmation
				}
				if reads == tc.failRead {
					return []byte(text), errors.New(secret) // Reader errors must not leak input.
				}
				return []byte(text), nil
			})
			defer clear(password)
			if tc.valid {
				if err != nil || string(password) != tc.password || reads != 2 {
					t.Fatalf("valid password was rejected or altered: %v", err)
				}
			} else if err == nil || len(password) != 0 {
				t.Fatal("invalid or unconfirmed password accepted")
			}
			if strings.Contains(output.String(), secret) || err != nil && strings.Contains(err.Error(), secret) {
				t.Error("password leaked in prompt or error")
			}
		})
	}
	if data, err := encodePFX(nil, nil, nil, nil); err == nil || len(data) != 0 {
		t.Fatal("encoding accepted an empty password")
	}
}

func TestPFXNonInteractiveDoesNotWrite(t *testing.T) {
	for _, language := range []string{"en", "pl"} {
		t.Run(language, func(t *testing.T) {
			root := t.TempDir()
			secret := "not-a-command-line-password"
			before := snapshotPEMTestTree(t, root)
			output, err := runCLIWithInput(t, root, secret+"\n"+secret+"\n", "-lang="+language, "-pkcs12", "localhost")
			if err == nil || !strings.Contains(string(output), "PFX") || strings.Contains(string(output), secret) {
				t.Fatalf("noninteractive PFX should fail without leaking input: %v\n%s", err, output)
			}
			if after := snapshotPEMTestTree(t, root); !reflect.DeepEqual(before, after) {
				t.Error("noninteractive PFX created CA or output files")
			}
		})
	}
	root := t.TempDir()
	path := filepath.Join(root, "existing.pfx")
	if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	before := snapshotPEMTestTree(t, root)
	output, err := runB04CLI(t, root, "-pkcs12", "-p12-file", path, "localhost")
	if err == nil || !strings.Contains(string(output), "refusing to overwrite existing PFX") {
		t.Fatalf("existing PFX should be refused before password/CA: %v\n%s", err, output)
	}
	if after := snapshotPEMTestTree(t, root); !reflect.DeepEqual(before, after) {
		t.Error("refused PFX operation changed files")
	}
}

func TestPFXGenerationAndReadback(t *testing.T) {
	for _, ecdsaKey := range []bool{false, true} {
		name := "rsa"
		if ecdsaKey {
			name = "ecdsa"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			// An ephemeral issuer, held only in memory, never installed or saved.
			caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				t.Fatal(err)
			}
			tpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "ephemeral test issuer"}, IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour)}
			der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &caKey.PublicKey, caKey)
			if err != nil {
				t.Fatal(err)
			}
			caCert, err := x509.ParseCertificate(der)
			if err != nil {
				t.Fatal(err)
			}
			password := []byte("c4a35f072d9a8e192ba1-strong-test-password")
			defer clear(password)
			path := filepath.Join(root, "server.pfx")
			m := &mkcert{pkcs12: true, ecdsa: ecdsaKey, pfxPassword: password, p12File: path, caKey: caKey, caCert: caCert}
			var logs bytes.Buffer
			originalOutput := log.Writer()
			log.SetOutput(&logs)
			t.Cleanup(func() { log.SetOutput(originalOutput) })
			m.makeCert([]string{"server.test"})
			if bytes.Contains(logs.Bytes(), password) || strings.Contains(logs.String(), "changeit") {
				t.Fatal("PFX log contains a password")
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			key, cert, chain, err := pkcs12.DecodeChain(data, string(password))
			if err != nil {
				t.Fatal(err)
			}
			if len(chain) != 1 || !chain[0].Equal(caCert) || cert.VerifyHostname("server.test") != nil {
				t.Fatal("PFX certificate or CA chain does not match")
			}
			if err := cert.CheckSignatureFrom(caCert); err != nil {
				t.Fatal(err)
			}
			public := key.(crypto.Signer).Public()
			pubDER, _ := x509.MarshalPKIXPublicKey(public)
			leafDER, _ := x509.MarshalPKIXPublicKey(cert.PublicKey)
			caDER, _ := x509.MarshalPKIXPublicKey(caCert.PublicKey)
			if !bytes.Equal(pubDER, leafDER) || bytes.Equal(pubDER, caDER) {
				t.Fatal("PFX must contain the server key, never the private CA key")
			}
			for _, wrong := range []string{"", "changeit", "wrong-password"} {
				if _, _, _, err := pkcs12.DecodeChain(data, wrong); err == nil {
					t.Fatal("PFX readable with an incorrect password")
				}
			}
			checkModernPFXAlgorithms(t, data)
			if runtime.GOOS != "windows" {
				info, err := os.Stat(path)
				if err != nil || info.Mode().Perm()&0077 != 0 {
					t.Fatal("PFX permissions are not private", err)
				}
			}
		})
	}
}

func TestPFXRefusesOverwrite(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "existing.pfx")
	if err := writePFXFile(path, []byte("original")); err != nil {
		t.Fatal(err)
	}
	before := snapshotPEMTestTree(t, root)
	if err := writePFXFile(path, []byte("replacement")); err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("expected refusal to overwrite: %v", err)
	}
	if after := snapshotPEMTestTree(t, root); !reflect.DeepEqual(before, after) {
		t.Error("existing PFX changed")
	}
	if err := writePFXFile(root, []byte("replacement")); err == nil {
		t.Fatal("directory accepted as PFX target")
	}
}

// Inspect the emitted ASN.1, independently of the selected encoder's name.
func checkModernPFXAlgorithms(t *testing.T, data []byte) {
	t.Helper()
	type contentInfo struct {
		Type    asn1.ObjectIdentifier
		Content asn1.RawValue `asn1:"tag:0,explicit"`
	}
	var pfx struct {
		Version  int
		AuthSafe contentInfo
		MAC      struct {
			Digest struct {
				Algorithm pkix.AlgorithmIdentifier
				Value     []byte
			}
			Salt       []byte
			Iterations int
		}
	}
	decode := func(data []byte, value any) {
		t.Helper()
		if rest, err := asn1.Unmarshal(data, value); err != nil || len(rest) != 0 {
			t.Fatalf("invalid PFX ASN.1: %v", err)
		}
	}
	decode(data, &pfx)
	if pfx.Version != 3 || pfx.MAC.Digest.Algorithm.Algorithm.String() != "2.16.840.1.101.3.4.2.1" || pfx.MAC.Iterations != 2048 || len(pfx.MAC.Salt) != 16 {
		t.Fatal("PFX does not use Modern2023 SHA-256 MAC parameters")
	}
	var safeBytes []byte
	decode(pfx.AuthSafe.Content.Bytes, &safeBytes)
	var safes []contentInfo
	decode(safeBytes, &safes)
	var algorithms []pkix.AlgorithmIdentifier
	for _, safe := range safes {
		switch safe.Type.String() {
		case "1.2.840.113549.1.7.6": // Encrypted certificates.
			var encrypted struct {
				Version int
				Content struct {
					Type      asn1.ObjectIdentifier
					Algorithm pkix.AlgorithmIdentifier
					Data      []byte `asn1:"tag:0"`
				}
			}
			decode(safe.Content.Bytes, &encrypted)
			algorithms = append(algorithms, encrypted.Content.Algorithm)
		case "1.2.840.113549.1.7.1": // Encrypted private key bag in a data safe.
			var bagsData []byte
			decode(safe.Content.Bytes, &bagsData)
			var bags []struct {
				ID         asn1.ObjectIdentifier
				Value      asn1.RawValue `asn1:"tag:0,explicit"`
				Attributes asn1.RawValue `asn1:"optional"`
			}
			decode(bagsData, &bags)
			if len(bags) != 1 || bags[0].ID.String() != "1.2.840.113549.1.12.10.1.2" {
				t.Fatal("expected exactly one encrypted private key")
			}
			var key struct {
				Algorithm pkix.AlgorithmIdentifier
				Data      []byte
			}
			decode(bags[0].Value.Bytes, &key)
			algorithms = append(algorithms, key.Algorithm)
		default:
			t.Fatal("unexpected PFX safe")
		}
	}
	if len(algorithms) != 2 {
		t.Fatal("missing encrypted certificate or private key")
	}
	for _, algorithm := range algorithms {
		if algorithm.Algorithm.String() != "1.2.840.113549.1.5.13" {
			t.Fatal("PFX encryption is not PBES2")
		}
		var params struct{ KDF, Cipher pkix.AlgorithmIdentifier }
		decode(algorithm.Parameters.FullBytes, &params)
		var kdf struct {
			Salt       []byte
			Iterations int
			KeyLength  int `asn1:"optional"`
			PRF        pkix.AlgorithmIdentifier
		}
		decode(params.KDF.Parameters.FullBytes, &kdf)
		if params.KDF.Algorithm.String() != "1.2.840.113549.1.5.12" || params.Cipher.Algorithm.String() != "2.16.840.1.101.3.4.1.42" || kdf.PRF.Algorithm.String() != "1.2.840.113549.2.9" || kdf.Iterations != 2048 || len(kdf.Salt) != 16 {
			t.Fatal("PFX does not use PBKDF2-HMAC-SHA-256/AES-256-CBC with Modern2023 parameters")
		}
	}
}
