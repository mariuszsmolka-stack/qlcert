package main

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Fatal CLI errors exit the process, so exercise main in an isolated child.
func TestB04CLIProcess(t *testing.T) {
	if os.Getenv("QUESTLAB_B04_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"mkcert"}, os.Args[i+1:]...)
			flag.CommandLine = flag.NewFlagSet("mkcert", flag.ExitOnError)
			main()
			os.Exit(0)
		}
	}
	t.Fatal("missing CLI argument separator")
}

func runB04CLI(t *testing.T, root string, args ...string) ([]byte, error) {
	t.Helper()
	return runCLIWithInput(t, root, "", args...)
}

func runCLIWithInput(t *testing.T, root, input string, args ...string) ([]byte, error) {
	t.Helper()
	t.Setenv("QUESTLAB_B04_HELPER", "1")
	t.Setenv("CAROOT", filepath.Join(root, "ca"))
	t.Setenv("HOME", filepath.Join(root, "user home"))
	t.Setenv("USERPROFILE", filepath.Join(root, "user home"))
	// Disable every trust store, including in a regression of the CLI ordering.
	t.Setenv("TRUST_STORES", "none")
	t.Setenv("JAVA_HOME", "")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(executable, append([]string{"-test.run=^TestB04CLIProcess$", "--"}, args...)...)
	cmd.Dir = root
	cmd.Stdin = strings.NewReader(input)
	return cmd.CombinedOutput()
}

func TestB04RejectedOperationsDoNotCreateCA(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		prepare    func(*testing.T, string)
		wantExit   int
		wantOutput string
	}{
		{
			name: "uninstall_missing_directory", args: []string{"-uninstall"},
			wantExit: 1, wantOutput: "failed to read the CA certificate",
		},
		{
			name: "uninstall_empty_directory", args: []string{"-uninstall"},
			prepare: func(t *testing.T, root string) {
				if err := os.Mkdir(filepath.Join(root, "ca"), 0700); err != nil {
					t.Fatal(err)
				}
			},
			wantExit: 1, wantOutput: "failed to read the CA certificate",
		},
		{
			name: "uninstall_preserves_orphan_key", args: []string{"-uninstall"},
			prepare: func(t *testing.T, root string) {
				if err := os.Mkdir(filepath.Join(root, "ca"), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, "ca", rootKeyName), []byte("existing key marker"), 0600); err != nil {
					t.Fatal(err)
				}
			},
			wantExit: 1, wantOutput: "failed to read the CA certificate",
		},
		{name: "invalid_dns", args: []string{"bad name"}, wantExit: 1, wantOutput: "not a valid hostname"},
		{name: "invalid_ipv4", args: []string{"192.0.2.1/24"}, wantExit: 1, wantOutput: "not a valid hostname"},
		{name: "invalid_ipv6", args: []string{"2001:db8::zz"}, wantExit: 1, wantOutput: "not a valid hostname"},
		{name: "mixed_names", args: []string{"localhost", "bad name"}, wantExit: 1, wantOutput: "not a valid hostname"},
		{name: "install_with_invalid_name", args: []string{"-install", "bad name"}, wantExit: 1, wantOutput: "not a valid hostname"},
		{name: "missing_names", args: []string{"-ecdsa"}, wantOutput: "QuestLab Cert"},
		{name: "unknown_flag", args: []string{"-unknown"}, wantExit: 2, wantOutput: "flag provided but not defined"},
		{name: "conflicting_flags", args: []string{"-install", "-uninstall"}, wantExit: 1, wantOutput: "can't set -install and -uninstall"},
		{name: "missing_csr", args: []string{"-csr", "missing.csr"}, wantExit: 1, wantOutput: "failed to read the CSR"},
		{
			name: "malformed_csr", args: []string{"-csr", "invalid.csr"},
			prepare: func(t *testing.T, root string) {
				if err := os.WriteFile(filepath.Join(root, "invalid.csr"), []byte("not a CSR"), 0600); err != nil {
					t.Fatal(err)
				}
			},
			wantExit: 1, wantOutput: "failed to read the CSR: unexpected content",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			if tt.prepare != nil {
				tt.prepare(t, root)
			}
			before := snapshotPEMTestTree(t, root)
			output, err := runB04CLI(t, root, tt.args...)
			if tt.wantExit == 0 {
				if err != nil {
					t.Fatalf("CLI failed: %v\n%s", err, output)
				}
			} else {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) || exitErr.ExitCode() != tt.wantExit {
					t.Fatalf("CLI error = %v, want exit %d\n%s", err, tt.wantExit, output)
				}
			}
			if !strings.Contains(string(output), tt.wantOutput) {
				t.Errorf("CLI output = %s, want %q", output, tt.wantOutput)
			}
			if after := snapshotPEMTestTree(t, root); !reflect.DeepEqual(after, before) {
				t.Errorf("operation created or changed files: before=%v, after=%v", before, after)
			}
		})
	}
}

func TestB04ValidGenerationAndCAReuse(t *testing.T) {
	root := t.TempDir()
	var initialCert, initialKey []byte
	for _, prefix := range []string{"first", "second"} {
		t.Run(prefix, func(t *testing.T) {
			output, err := runB04CLI(t, root, "-ecdsa", "-cert-file", prefix+".pem", "-key-file", prefix+"-key.pem", "localhost", "127.0.0.1", "::1")
			if err != nil {
				t.Fatalf("generation failed: %v\n%s", err, output)
			}
			caBytes, err := os.ReadFile(filepath.Join(root, "ca", rootName))
			if err != nil {
				t.Fatal(err)
			}
			caKeyBytes, err := os.ReadFile(filepath.Join(root, "ca", rootKeyName))
			if err != nil {
				t.Fatal(err)
			}
			if initialCert == nil {
				initialCert, initialKey = caBytes, caKeyBytes
			} else if !bytes.Equal(caBytes, initialCert) || !bytes.Equal(caKeyBytes, initialKey) {
				t.Error("existing test CA was replaced")
			}
			ca := parseB04Certificate(t, caBytes)
			leafBytes, err := os.ReadFile(filepath.Join(root, prefix+".pem"))
			if err != nil {
				t.Fatal(err)
			}
			leaf := parseB04Certificate(t, leafBytes)
			roots := x509.NewCertPool()
			roots.AddCert(ca)
			if _, err := leaf.Verify(x509.VerifyOptions{Roots: roots, DNSName: "localhost"}); err != nil {
				t.Errorf("generated certificate does not verify against test CA: %v", err)
			}
			if len(leaf.IPAddresses) != 2 || !leaf.IPAddresses[0].Equal(net.ParseIP("127.0.0.1")) || !leaf.IPAddresses[1].Equal(net.ParseIP("::1")) {
				t.Errorf("unexpected IP SANs: %v", leaf.IPAddresses)
			}
			keyBytes, err := os.ReadFile(filepath.Join(root, prefix+"-key.pem"))
			if err != nil {
				t.Fatal(err)
			}
			block, _ := pem.Decode(keyBytes)
			if block == nil {
				t.Fatal("missing PEM key")
			}
			key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				t.Fatal(err)
			}
			signer, ok := key.(crypto.Signer)
			if !ok {
				t.Fatalf("unexpected key type %T", key)
			}
			publicKey, err := x509.MarshalPKIXPublicKey(signer.Public())
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(publicKey, leaf.RawSubjectPublicKeyInfo) {
				t.Error("generated private key does not match certificate")
			}
		})
	}
}

func parseB04Certificate(t *testing.T, data []byte) *x509.Certificate {
	t.Helper()
	block, _ := pem.Decode(data)
	if block == nil {
		t.Fatal("missing PEM certificate")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	return cert
}
