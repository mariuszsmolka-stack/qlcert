package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestWriteCertAndKeyPEM(t *testing.T) {
	tests := []struct {
		name          string
		certName      string
		keyName       string
		existingFiles map[string]string
		existingDirs  []string
		wantErr       bool
		wantExists    bool
	}{
		{name: "new_pair", certName: "cert.pem", keyName: "key.pem"},
		{
			name: "existing_certificate", certName: "cert.pem", keyName: "key.pem",
			existingFiles: map[string]string{"cert.pem": "original certificate"}, wantErr: true, wantExists: true,
		},
		{
			name: "existing_key", certName: "cert.pem", keyName: "key.pem",
			existingFiles: map[string]string{"key.pem": "original private key"}, wantErr: true, wantExists: true,
		},
		{
			name: "both_exist", certName: "cert.pem", keyName: "key.pem",
			existingFiles: map[string]string{"cert.pem": "original certificate", "key.pem": "original private key"},
			wantErr:       true, wantExists: true,
		},
		{
			name: "second_parent_missing", certName: "cert.pem", keyName: "missing/key.pem",
			wantErr: true,
		},
		{
			name: "second_parent_is_file", certName: "cert.pem", keyName: "parent/key.pem",
			existingFiles: map[string]string{"parent": "keep this file"}, wantErr: true,
		},
		{
			name: "second_target_is_directory", certName: "cert.pem", keyName: "key.pem",
			existingDirs: []string{"key.pem"}, wantErr: true,
		},
		{
			name: "first_parent_missing", certName: "missing/cert.pem", keyName: "key.pem",
			wantErr: true,
		},
		{name: "new_combined", certName: "combined.pem", keyName: "combined.pem"},
		{
			name: "existing_combined", certName: "combined.pem", keyName: "combined.pem",
			existingFiles: map[string]string{"combined.pem": "original combined PEM"}, wantErr: true, wantExists: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			for _, name := range tt.existingDirs {
				if err := os.Mkdir(filepath.Join(root, name), 0700); err != nil {
					t.Fatal(err)
				}
			}
			for name, data := range tt.existingFiles {
				if err := os.WriteFile(filepath.Join(root, name), []byte(data), 0644); err != nil {
					t.Fatal(err)
				}
			}
			before := snapshotPEMTestTree(t, root)
			certPath, keyPath := filepath.Join(root, tt.certName), filepath.Join(root, tt.keyName)
			err := writeCertAndKeyPEM(certPath, keyPath, []byte("certificate data\n"), []byte("private key data\n"))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected a write error")
				}
				if tt.wantExists && (!errors.Is(err, fs.ErrExist) || !strings.Contains(err.Error(), "refusing to overwrite")) {
					t.Errorf("expected a clear existing-file error, got %v", err)
				}
				if !strings.Contains(err.Error(), certPath) && !strings.Contains(err.Error(), keyPath) {
					t.Errorf("error does not identify a target: %v", err)
				}
				if after := snapshotPEMTestTree(t, root); !reflect.DeepEqual(after, before) {
					t.Errorf("failed operation changed existing files or left new files: before=%v, after=%v", before, after)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			wantCert := "certificate data\n"
			if certPath == keyPath {
				wantCert += "private key data\n"
			}
			assertPEMTestContents(t, certPath, wantCert)
			if certPath != keyPath {
				assertPEMTestContents(t, keyPath, "private key data\n")
			}
			if runtime.GOOS != "windows" {
				info, err := os.Stat(keyPath)
				if err != nil {
					t.Fatal(err)
				}
				if info.Mode().Perm()&0077 != 0 {
					t.Errorf("private key permissions = %o, allow group/other access", info.Mode().Perm())
				}
			}
		})
	}
}

func TestWritePEMFilesSingleCertificate(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "csr-cert.pem")
	if err := writePEMFiles(pemFile{path, []byte("first certificate"), 0644}); err != nil {
		t.Fatal(err)
	}
	err := writePEMFiles(pemFile{path, []byte("replacement certificate"), 0644})
	if !errors.Is(err, fs.ErrExist) || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Errorf("expected refusal to overwrite CSR certificate, got %v", err)
	}
	assertPEMTestContents(t, path, "first certificate")
}

func TestWritePEMFilesConcurrentCreation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cert.pem")
	type result struct {
		data string
		err  error
	}
	const writers = 8
	start := make(chan struct{})
	results := make(chan result, writers)
	for i := 0; i < writers; i++ {
		data := fmt.Sprintf("certificate from writer %d", i)
		go func() {
			<-start
			results <- result{data, writePEMFiles(pemFile{path, []byte(data), 0644})}
		}()
	}
	close(start)
	successes := 0
	var winner string
	for i := 0; i < writers; i++ {
		r := <-results
		if r.err == nil {
			successes++
			winner = r.data
		} else if !errors.Is(r.err, fs.ErrExist) {
			t.Errorf("unexpected concurrent creation error: %v", r.err)
		}
	}
	if successes != 1 {
		t.Fatalf("successful writers = %d, want 1", successes)
	}
	assertPEMTestContents(t, path, winner)
}

func assertPEMTestContents(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != want {
		t.Errorf("contents of %q = %q, want %q", path, data, want)
	}
}

type pemTestEntry struct {
	mode    os.FileMode
	data    string
	modTime time.Time
}

func snapshotPEMTestTree(t *testing.T, root string) map[string]pemTestEntry {
	t.Helper()
	snapshot := make(map[string]pemTestEntry)
	err := filepath.WalkDir(root, func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		info, err := dirEntry.Info()
		if err != nil {
			return err
		}
		item := pemTestEntry{mode: info.Mode()}
		if !dirEntry.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			item.data = string(data)
			item.modTime = info.ModTime()
		}
		snapshot[rel] = item
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}
