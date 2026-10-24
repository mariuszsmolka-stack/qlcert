package main

import (
	"bufio"
	"bytes"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestHTTPSInstructionSelection(t *testing.T) {
	originalLanguage := uiLanguage
	t.Cleanup(func() { uiLanguage = originalLanguage })
	for _, language := range []string{"en", "pl"} {
		for _, tc := range []struct {
			name, input, host string
			want              []string
		}{
			{"apache", "1\n4\n", "lab.test", []string{"Apache 2.4 HTTPS", "SSLCertificateKeyFile /etc/ssl/private/questlab-cert-key.pem"}},
			{"nginx", "2\n4\n", "lab.test", []string{"Nginx HTTPS", "listen 443 ssl;", "server_name lab.test;", "ssl_certificate /etc/ssl/certs/questlab-cert.pem;", "ssl_certificate_key /etc/ssl/private/questlab-cert-key.pem;", "-m 600 server-key.pem /etc/ssl/private/questlab-cert-key.pem", "sudo nginx -t && sudo systemctl reload nginx"}},
			{"iis_dns", "3\n4\n", "lab.test", []string{"IIS HTTPS", "Modern2023", "changeit", "Import-PfxCertificate -FilePath 'C:\\Certificates\\server.pfx'", "Cert:\\LocalMachine\\My", "-AsSecureString", "certlm.msc", "inetmgr", "SNI", "=lab.test"}},
			{"iis_ip", "3\n4\n", "192.0.2.1", []string{"IIS HTTPS", "IP 192.0.2.1", "IP:443"}},
			{"nginx_ipv6", "2\n4\n", "2001:db8::1", []string{"listen [::]:443 ssl;", "https://[2001:db8::1]/"}},
			{"multiple_and_invalid", "9\n1\n2\n3\n4\n", "lab.test", []string{"1, 2, 3", "Apache 2.4 HTTPS", "Nginx HTTPS", "IIS HTTPS"}},
		} {
			t.Run(language+"_"+tc.name, func(t *testing.T) {
				uiLanguage = language
				root := t.TempDir()
				certPath := filepath.Join(root, "server certificate.pem")
				keyPath := filepath.Join(root, "server private key.pem")
				caPath := filepath.Join(root, "ca", rootName)
				before := snapshotPEMTestTree(t, root)
				var output bytes.Buffer
				if err := showHTTPSInstructions(bufio.NewScanner(strings.NewReader(tc.input)), &output, tc.host, certPath, keyPath, caPath); err != nil {
					t.Fatal(err)
				}
				common := []string{certPath, keyPath, caPath, "rootCA.pem", "rootCA-key.pem", "certmgr.msc", "sudo update-ca-certificates"}
				if language == "en" {
					common = append(common, "4. Back", "manual steps", "NEVER copy or share rootCA-key.pem", "does not make browsers trust it", "localhost refers to the client itself")
					if strings.Contains(tc.input, "3\n") {
						common = append(common, "Do not use older PFX exports", "legacy algorithms", "import into real IIS has not been verified", "Continue ONLY with that separately prepared PFX")
					}
				} else {
					common = append(common, "4. Powrót", "kroki ręczne", "NIGDY nie kopiuj ani nie udostępniaj rootCA-key.pem", "nie oznacza zaufania przeglądarki", "localhost oznacza sam komputer klienta")
					if strings.Contains(tc.input, "3\n") {
						common = append(common, "Nie używaj dawnych eksportów PFX", "starszymi algorytmami", "importu w rzeczywistym IIS nie zweryfikowano", "Kontynuuj WYŁĄCZNIE z tak przygotowanym osobno PFX")
					}
				}
				for _, want := range append(common, tc.want...) {
					if !strings.Contains(output.String(), want) {
						t.Errorf("instructions omit %q: %s", want, &output)
					}
				}
				if strings.Contains(output.String(), "%!") {
					t.Error("invalid instruction formatting")
				}
				if after := snapshotPEMTestTree(t, root); !reflect.DeepEqual(before, after) {
					t.Error("displaying instructions wrote files")
				}
			})
		}
	}
}

func TestHTTPSInstructionsBackAndEOF(t *testing.T) {
	originalLanguage := uiLanguage
	t.Cleanup(func() { uiLanguage = originalLanguage })
	for _, language := range []string{"en", "pl"} {
		for _, input := range []string{"4\n", ""} {
			t.Run(language+"_"+input, func(t *testing.T) {
				uiLanguage = language
				var output bytes.Buffer
				if err := showHTTPSInstructions(bufio.NewScanner(strings.NewReader(input)), &output, "lab.test", "server.pem", "server-key.pem", "rootCA.pem"); err != nil {
					t.Fatal(err)
				}
				if strings.Contains(output.String(), "manual steps") || strings.Contains(output.String(), "kroki ręczne") {
					t.Error("instructions displayed without selection")
				}
			})
		}
	}
}

func TestIISInstructionsForGeneratedPFX(t *testing.T) {
	originalLanguage := uiLanguage
	t.Cleanup(func() { uiLanguage = originalLanguage })
	for _, language := range []string{"en", "pl"} {
		t.Run(language, func(t *testing.T) {
			uiLanguage = language
			root := t.TempDir()
			path := filepath.Join(root, "user's home", "QuestLab-Cert", "certificates", "server.pfx")
			var output bytes.Buffer
			printIISInstructions(&output, "server.test", "", "", filepath.Join(root, "ca", rootName), path)
			for _, want := range []string{path, "-FilePath '" + strings.ReplaceAll(path, "'", "''") + "'", "Modern2023", "-AsSecureString", "NEVER copy"} {
				if want == "NEVER copy" && language == "pl" {
					want = "NIGDY nie kopiuj"
				}
				if !strings.Contains(output.String(), want) {
					t.Errorf("PFX instructions omit %q: %s", want, &output)
				}
			}
			warning := "Import into real IIS has not yet been tested"
			if language == "pl" {
				warning = "Import na rzeczywistym IIS nie został jeszcze przetestowany"
			}
			if !strings.Contains(output.String(), warning) || strings.Contains(output.String(), `C:\Certificates\server.pfx`) || strings.Contains(output.String(), "%!") {
				t.Error("PFX instructions contain a placeholder path, missing warning or formatting error")
			}
		})
	}
}
