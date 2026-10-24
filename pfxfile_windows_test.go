package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func TestPFXWindowsACL(t *testing.T) {
	dir := t.TempDir()
	// A deliberately broad parent ACL must not leak into the new PFX file.
	sd, err := windows.SecurityDescriptorFromString("D:P(A;OICI;FA;;;WD)")
	if err != nil {
		t.Fatal(err)
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		t.Fatal(err)
	}
	if err := windows.SetNamedSecurityInfo(dir, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "private.pfx")
	if err := writePFXFile(path, []byte("private bundle")); err != nil {
		t.Fatal(err)
	}
	actual, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(actual.String(), "D:P(A;;FA;;;"+user.User.Sid.String()+")") || strings.Contains(actual.String(), ";;;WD)") {
		t.Fatalf("PFX inherited broad permissions: %s", actual.String())
	}
	unsafePath := filepath.Join(dir, "inherited.pfx")
	if err := os.WriteFile(unsafePath, []byte("marker"), 0600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(unsafePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := verifyPFXPermissions(f); err == nil {
		t.Fatal("PFX verification accepted an inherited Everyone ACL")
	}
}
