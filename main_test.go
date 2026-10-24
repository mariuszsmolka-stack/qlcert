package main

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestStoreEnabled(t *testing.T) {
	tests := []struct {
		name   string
		stores string
		unset  bool
		store  string
		want   bool
	}{
		{name: "unset_system", unset: true, store: "system", want: true},
		{name: "unset_nss", unset: true, store: "nss", want: true},
		{name: "unset_java", unset: true, store: "java", want: true},
		{name: "empty_configuration", stores: "", store: "system", want: true},
		{name: "single_selected", stores: "system", store: "system", want: true},
		{name: "single_unlisted", stores: "system", store: "java", want: false},
		{name: "multiple_system", stores: "system,nss", store: "system", want: true},
		{name: "multiple_nss", stores: "system,nss", store: "nss", want: true},
		{name: "multiple_unlisted", stores: "system,nss", store: "java", want: false},
		{name: "exact_token", stores: "system,nss", store: "sys", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Register restoration with Setenv even when testing an absent variable.
			t.Setenv("TRUST_STORES", tt.stores)
			if tt.unset {
				if err := os.Unsetenv("TRUST_STORES"); err != nil {
					t.Fatal(err)
				}
			}
			if got := storeEnabled(tt.store); got != tt.want {
				t.Errorf("storeEnabled(%q) = %t, want %t", tt.store, got, tt.want)
			}
		})
	}
}

func TestGetCAROOT(t *testing.T) {
	tests := []struct {
		name     string
		explicit bool
		xdg      bool
	}{
		{name: "explicit_caroot", explicit: true, xdg: true},
		{name: "platform_default"},
		{name: "xdg_data_home", xdg: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Every candidate location is inside a temporary directory. The
			// candidates themselves do not exist, so unintended creation is visible.
			root := t.TempDir()
			home := filepath.Join(root, "home")
			localAppData := filepath.Join(root, "local-app-data")
			xdgDataHome := filepath.Join(root, "xdg-data")
			t.Setenv("CAROOT", "")
			t.Setenv("HOME", home)
			t.Setenv("LocalAppData", localAppData)
			t.Setenv("XDG_DATA_HOME", "")
			if tt.xdg {
				t.Setenv("XDG_DATA_HOME", xdgDataHome)
			}
			if tt.explicit {
				t.Setenv("CAROOT", filepath.Join(root, "explicit-ca"))
			}

			var want string
			switch {
			case tt.explicit:
				want = filepath.Join(root, "explicit-ca")
			case runtime.GOOS == "windows":
				want = filepath.Join(localAppData, "mkcert")
			case tt.xdg:
				want = filepath.Join(xdgDataHome, "mkcert")
			case runtime.GOOS == "darwin":
				want = filepath.Join(home, "Library", "Application Support", "mkcert")
			default:
				want = filepath.Join(home, ".local", "share", "mkcert")
			}

			beforeCall := os.Environ()
			if got := getCAROOT(); got != want {
				t.Errorf("getCAROOT() on %s = %q, want %q", runtime.GOOS, got, want)
			}
			if !reflect.DeepEqual(os.Environ(), beforeCall) {
				t.Error("getCAROOT() modified the process environment")
			}
			entries, err := os.ReadDir(root)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Errorf("getCAROOT() created files or directories: %v", entries)
			}
		})
	}
}

func TestPathExists(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "existing.txt")
	if err := os.WriteFile(file, []byte("test fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "existing-directory")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "existing_file", path: file, want: true},
		{name: "existing_directory", path: dir, want: true},
		{name: "missing_path", path: filepath.Join(root, "missing"), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathExists(tt.path); got != tt.want {
				t.Errorf("pathExists(%q) = %t, want %t", tt.path, got, tt.want)
			}
		})
	}
}
