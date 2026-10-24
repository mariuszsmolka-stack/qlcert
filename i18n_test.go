package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestCLILanguages(t *testing.T) {
	root := t.TempDir()
	fullHelpFlags := []string{"-help", "-version", "-lang", "-install", "-uninstall", "-CAROOT", "-cert-file", "-key-file", "-p12-file", "-client", "-ecdsa", "-pkcs12", "-csr"}
	tests := []struct {
		name     string
		args     []string
		want     string
		wantExit int
		fullHelp bool
	}{
		{name: "default_help", args: []string{"-help"}, want: "Options:", fullHelp: true},
		{name: "english_help", args: []string{"-lang=en", "-help"}, want: "Options:", fullHelp: true},
		{name: "polish_help", args: []string{"-lang=pl", "-help"}, want: "Opcje:", fullHelp: true},
		{name: "language_after_help", args: []string{"-help", "-lang=pl"}, want: "Opcje:", fullHelp: true},
		{name: "no_arguments", want: "1. Generate HTTPS certificate"},
		{name: "polish_only", args: []string{"-lang=pl"}, want: "1. Utwórz certyfikat HTTPS"},
		{name: "english_only", args: []string{"-lang=en"}, want: "1. Generate HTTPS certificate"},
		{name: "short_help_alias", args: []string{"-lang=pl", "-h"}, want: "Użycie:"},
		{name: "english_version", args: []string{"-version"}, want: "Based on mkcert"},
		{name: "polish_version", args: []string{"-lang=pl", "-version"}, want: "Na podstawie mkcert"},
		{name: "unsupported_language", args: []string{"-lang=de", "-help"}, want: "use -lang=en or -lang=pl", wantExit: 1},
		{name: "empty_language", args: []string{"-lang="}, want: "use -lang=en or -lang=pl", wantExit: 1},
		{name: "uppercase_language", args: []string{"-lang=PL"}, want: "use -lang=en or -lang=pl", wantExit: 1},
		{name: "polish_unknown_flag", args: []string{"-lang=pl", "-unknown"}, want: "nieznany parametr: -unknown", wantExit: 2},
		{name: "polish_missing_value", args: []string{"-lang=pl", "-cert-file"}, want: "parametr wymaga wartości: -cert-file", wantExit: 2},
		{name: "polish_bad_flag_syntax", args: []string{"-lang=pl", "---bad"}, want: "nieprawidłowa składnia parametru: ---bad", wantExit: 2},
		{name: "polish_invalid_boolean", args: []string{"-lang=pl", "-install=invalid"}, want: "nieprawidłowa wartość logiczna \"invalid\" dla parametru -install: błąd składni", wantExit: 2},
		{name: "polish_conflict", args: []string{"-lang=pl", "-install", "-uninstall"}, want: "BŁĄD: nie można jednocześnie", wantExit: 1},
		{name: "polish_invalid_name", args: []string{"-lang=pl", "bad name"}, want: "nie jest poprawną nazwą hosta", wantExit: 1},
	}
	outputs := make(map[string]string)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Locale variables must not change the default or explicit language.
			t.Setenv("LANG", "pl_PL.UTF-8")
			t.Setenv("LC_ALL", "pl_PL.UTF-8")
			t.Setenv("LANGUAGE", "pl")
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
			text := string(output)
			if !strings.Contains(text, tt.want) {
				t.Errorf("output = %s, want %q", output, tt.want)
			}
			if tt.wantExit == 0 && !strings.Contains(text, "QuestLab Cert") {
				t.Error("program name missing")
			}
			if tt.fullHelp {
				for _, option := range fullHelpFlags {
					if !strings.Contains(text, option) {
						t.Errorf("full help omits %s", option)
					}
				}
				for _, notice := range []string{"mkcert", "The mkcert Authors", "BSD-3-Clause", "CAROOT", "TRUST_STORES", "JAVA_HOME"} {
					if !strings.Contains(text, notice) {
						t.Errorf("full help omits %q", notice)
					}
				}
			}
			if entries, err := os.ReadDir(root); err != nil || len(entries) != 0 {
				t.Fatalf("read-only CLI call wrote files: %v, %v", entries, err)
			}
			outputs[tt.name] = text
		})
	}
	if outputs["default_help"] != outputs["english_help"] {
		t.Error("default and explicit English help differ")
	}
	output, err := runB04CLI(t, root, "-help")
	if err != nil || string(output) != outputs["default_help"] {
		t.Errorf("language persisted across invocations: %v\n%s", err, output)
	}
	output, err = runB04CLI(t, root, "-lang=pl", "-CAROOT")
	if err != nil || strings.TrimSpace(string(output)) != filepath.Join(root, "ca") {
		t.Errorf("-CAROOT output changed: %v\n%s", err, output)
	}
}

func TestCLILanguagePreservesCA(t *testing.T) {
	root := t.TempDir()
	var originalCA, originalKey []byte
	for _, tc := range []struct{ language, message string }{
		{"en", "Created a new certificate"}, {"pl", "Utworzono nowy certyfikat"},
	} {
		t.Run(tc.language, func(t *testing.T) {
			output, err := runB04CLI(t, root, "-lang="+tc.language, "-ecdsa", "-client", "-cert-file", tc.language+".pem", "-key-file", tc.language+"-key.pem", "localhost")
			if err != nil || !strings.Contains(string(output), tc.message) || strings.Contains(string(output), "%!") {
				t.Fatalf("generation: %v\n%s", err, output)
			}
			ca, err := os.ReadFile(filepath.Join(root, "ca", rootName))
			if err != nil {
				t.Fatal(err)
			}
			key, err := os.ReadFile(filepath.Join(root, "ca", rootKeyName))
			if err != nil {
				t.Fatal(err)
			}
			if originalCA == nil {
				originalCA, originalKey = ca, key
			} else if !bytes.Equal(ca, originalCA) || !bytes.Equal(key, originalKey) {
				t.Error("language selection changed existing CA files")
			}
			cert := parseB04Certificate(t, ca)
			if !reflect.DeepEqual(cert.Subject.Organization, []string{"mkcert development CA"}) {
				t.Error("legacy CA metadata changed")
			}
			for _, path := range []string{tc.language + ".pem", tc.language + "-key.pem"} {
				if _, err := os.Stat(filepath.Join(root, path)); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestTranslationFormatArguments(t *testing.T) {
	verbs := regexp.MustCompile(`%[a-zA-Z%]`)
	for english, polish := range polishMessages {
		if polish == "" || !reflect.DeepEqual(verbs.FindAllString(english, -1), verbs.FindAllString(polish, -1)) {
			t.Errorf("translation changes format arguments: %q -> %q", english, polish)
		}
	}
}
