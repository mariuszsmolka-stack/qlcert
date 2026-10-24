package main

import (
	"bytes"
	"crypto/x509"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestMenuDoesNotWriteBeforeConfirmation(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		input    string
		want     string
		existing bool
	}{
		{name: "english_exit", input: "0\n", want: "0. Exit"},
		{name: "polish_exit", args: []string{"-lang=pl"}, input: "0\n", want: "0. Zakończ"},
		{name: "explicit_english", args: []string{"-lang=en"}, input: "0\n", want: "1. Generate HTTPS certificate"},
		{name: "exit_alias", input: "3\n", want: "0. Exit"},
		{name: "eof", want: "Choose an option:"},
		{name: "english_help", input: "2\n0\n", want: "Options:"},
		{name: "polish_help", args: []string{"-lang=pl"}, input: "2\n0\n", want: "Opcje:"},
		{name: "invalid_option", input: "9\n0\n", want: "Invalid option"},
		{name: "empty_name", input: "1\n\n0\n", want: "Cancelled."},
		{name: "invalid_dns", input: "1\nbad..example\n0\n", want: "Invalid input."},
		{name: "invalid_ipv4", input: "1\n999.0.0.1\n0\n", want: "Invalid input."},
		{name: "invalid_ipv6", input: "1\n2001:db8::zz\n0\n", want: "Invalid input."},
		{name: "multiple_names", input: "1\none.test two.test\n0\n", want: "Invalid input."},
		{name: "email_is_not_dns", input: "1\nuser@example.test\n0\n", want: "Invalid input."},
		{name: "url_is_not_dns", input: "1\nhttps://example.test\n0\n", want: "Invalid input."},
		{name: "cancel", input: "1\nlocalhost\nn\n0\n", want: "Cancelled."},
		{name: "default_no", input: "1\nlocalhost\n\n0\n", want: "Cancelled."},
		{name: "eof_at_name", input: "1\n", want: "Enter one DNS"},
		{name: "eof_at_confirmation", input: "1\nlocalhost\n", want: "Generate the certificate?"},
		{name: "ipv4_summary", input: "1\n192.0.2.1\nn\n0\n", want: "DNS/IP: 192.0.2.1"},
		{name: "ipv6_summary", input: "1\n2001:db8::1\nn\n0\n", want: "2001_db8__1.pem"},
		{name: "international_dns", input: "1\nbücher.example\nn\n0\n", want: "DNS/IP: xn--bcher-kva.example"},
		{name: "polish_cancel_existing", args: []string{"-lang=pl"}, input: "1\nlocalhost\nnie\n0\n", want: "Anulowano.", existing: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			if tt.existing {
				if err := os.Mkdir(filepath.Join(root, "ca"), 0700); err != nil {
					t.Fatal(err)
				}
				for _, name := range []string{rootName, rootKeyName} {
					if err := os.WriteFile(filepath.Join(root, "ca", name), []byte("existing CA marker"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			before := snapshotPEMTestTree(t, root)
			output, err := runCLIWithInput(t, root, tt.input, tt.args...)
			if err != nil {
				t.Fatalf("menu failed: %v\n%s", err, output)
			}
			if !strings.Contains(string(output), tt.want) {
				t.Errorf("menu output = %s, want %q", output, tt.want)
			}
			if strings.Contains(string(output), "Apache 2.4") {
				t.Error("Apache instructions offered before successful generation")
			}
			if strings.Contains(string(output), "Format: PEM") {
				dir := filepath.Join(root, "user home", "QuestLab-Cert", "certificates")
				if !filepath.IsAbs(dir) || strings.Count(string(output), dir+string(os.PathSeparator)) != 2 {
					t.Errorf("summary must show both absolute output paths: %s", output)
				}
			}
			if after := snapshotPEMTestTree(t, root); !reflect.DeepEqual(after, before) {
				t.Errorf("menu wrote files before confirmation: before=%v, after=%v", before, after)
			}
		})
	}
}

func TestMenuConfirmedPEMGeneration(t *testing.T) {
	root := t.TempDir()
	certDir := filepath.Join(root, "user home", "QuestLab-Cert", "certificates")
	var originalCert, originalKey []byte
	for _, tc := range []struct{ name, language, answer, summary, generated, guideAnswer, guideTitle string }{
		{"english_guide", "en", "yes", "Summary", "Created a new certificate", "yes\n0\n", "manual steps"},
		{"polish_guide", "pl", "tak", "Podsumowanie", "Utworzono nowy certyfikat", "tak\n0\n", "kroki ręczne"},
		{"decline_guide", "en", "yes", "Summary", "Created a new certificate", "n\n0\n", ""},
		{"eof_at_guide", "pl", "tak", "Podsumowanie", "Utworzono nowy certyfikat", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			host := strings.ReplaceAll(tc.name, "_", "-") + ".example.test"
			input := "1\n" + host + "\n" + tc.answer + "\n" + tc.guideAnswer
			var args []string
			if tc.language == "pl" {
				args = []string{"-lang=pl"}
			}
			output, err := runCLIWithInput(t, root, input, args...)
			if err != nil {
				t.Fatalf("menu generation failed: %v\n%s", err, output)
			}
			text := string(output)
			if summary, generated := strings.Index(text, tc.summary), strings.Index(text, tc.generated); summary < 0 || generated <= summary {
				t.Errorf("summary must precede generation: %s", output)
			}
			if !strings.Contains(text, "Format: PEM") {
				t.Error("PEM format missing from summary")
			}
			if prompt := strings.Index(text, "Apache 2.4"); prompt <= strings.Index(text, tc.generated) {
				t.Errorf("Apache instructions must be offered after generation: %s", output)
			}
			if tc.guideTitle != "" {
				for _, want := range []string{tc.guideTitle, "ServerName " + host, "https://" + host + "/", filepath.Join(root, "ca", rootName)} {
					if !strings.Contains(text, want) {
						t.Errorf("Apache instructions omit %q: %s", want, output)
					}
				}
			} else if strings.Contains(text, "<VirtualHost") {
				t.Error("Apache instructions displayed without confirmation")
			}
			certPath := filepath.Join(certDir, host+".pem")
			keyPath := filepath.Join(certDir, host+"-key.pem")
			for _, path := range []string{certPath, keyPath} {
				if !filepath.IsAbs(path) || !strings.Contains(text, path) {
					t.Errorf("absolute output path missing: %q\n%s", path, output)
				}
				if pathExists(filepath.Join(root, filepath.Base(path))) {
					t.Errorf("menu wrote a certificate or key into the working directory: %q", path)
				}
			}
			caBytes, err := os.ReadFile(filepath.Join(root, "ca", rootName))
			if err != nil {
				t.Fatal(err)
			}
			keyBytes, err := os.ReadFile(filepath.Join(root, "ca", rootKeyName))
			if err != nil {
				t.Fatal(err)
			}
			if originalCert == nil {
				originalCert, originalKey = caBytes, keyBytes
			} else if !bytes.Equal(originalCert, caBytes) || !bytes.Equal(originalKey, keyBytes) {
				t.Error("existing test CA was replaced")
			}
			certBytes, err := os.ReadFile(certPath)
			if err != nil {
				t.Fatal(err)
			}
			cert := parseB04Certificate(t, certBytes)
			roots := x509.NewCertPool()
			roots.AddCert(parseB04Certificate(t, caBytes))
			if _, err := cert.Verify(x509.VerifyOptions{Roots: roots, DNSName: host}); err != nil {
				t.Errorf("menu certificate does not verify: %v", err)
			}
			if _, err := os.Stat(keyPath); err != nil {
				t.Error(err)
			}
			for _, extension := range []string{".p12", ".pfx"} {
				if pathExists(filepath.Join(certDir, host+extension)) {
					t.Errorf("menu unexpectedly exported %s", extension)
				}
			}
		})
	}
}

func TestMenuHomeDirectoryPreservesExistingPEM(t *testing.T) {
	for _, existingName := range []string{"existing.test.pem", "existing.test-key.pem"} {
		t.Run(existingName, func(t *testing.T) {
			root := t.TempDir()
			// Prepare only a temporary CA so the check also verifies that a failed
			// output operation does not replace an existing CA or its key.
			if output, err := runB04CLI(t, root, "-ecdsa", "seed.test"); err != nil {
				t.Fatalf("test CA setup failed: %v\n%s", err, output)
			}
			dir := filepath.Join(root, "user home", "QuestLab-Cert", "certificates")
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, existingName), []byte("original PEM marker"), 0600); err != nil {
				t.Fatal(err)
			}
			before := snapshotPEMTestTree(t, root)
			output, err := runCLIWithInput(t, root, "1\nexisting.test\nyes\n0\n")
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 || !strings.Contains(string(output), "refusing to overwrite") {
				t.Fatalf("expected refusal to overwrite: %v\n%s", err, output)
			}
			if strings.Contains(string(output), "Apache 2.4") {
				t.Error("Apache instructions offered after failed generation")
			}
			if after := snapshotPEMTestTree(t, root); !reflect.DeepEqual(after, before) {
				t.Errorf("existing files changed or new files remained: before=%v, after=%v", before, after)
			}
		})
	}
}

func TestCLIDefaultOutputStaysInWorkingDirectory(t *testing.T) {
	root := t.TempDir()
	if output, err := runB04CLI(t, root, "-ecdsa", "cli.test"); err != nil {
		t.Fatalf("CLI generation failed: %v\n%s", err, output)
	} else if strings.Contains(string(output), "Apache 2.4") {
		t.Error("traditional CLI unexpectedly offers Apache instructions")
	}
	for _, name := range []string{"cli.test.pem", "cli.test-key.pem"} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Error(err)
		}
	}
	if pathExists(filepath.Join(root, "user home")) {
		t.Error("traditional CLI created the menu output directory")
	}
}
