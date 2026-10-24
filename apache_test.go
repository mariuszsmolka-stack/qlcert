package main

import (
	"bytes"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestApacheInstructions(t *testing.T) {
	originalLanguage := uiLanguage
	t.Cleanup(func() { uiLanguage = originalLanguage })
	for _, tc := range []struct{ language, host, serverName, warning, caWarning string }{
		{"en", "lab.example.test", "lab.example.test", "localhost refers to the client itself, not a remote server", "NEVER copy or share rootCA-key.pem"},
		{"pl", "192.0.2.10", "192.0.2.10", "localhost oznacza sam komputer klienta, a nie zdalny serwer", "NIGDY nie kopiuj ani nie udostępniaj rootCA-key.pem"},
		{"en", "2001:db8::1", "[2001:db8::1]", "localhost refers to the client itself, not a remote server", "NEVER copy or share rootCA-key.pem"},
	} {
		t.Run(tc.language+"_"+tc.host, func(t *testing.T) {
			uiLanguage = tc.language
			root := t.TempDir()
			certPath := filepath.Join(root, "server.pem")
			keyPath := filepath.Join(root, "server-key.pem")
			caPath := filepath.Join(root, "ca", rootName)
			before := snapshotPEMTestTree(t, root)
			var output bytes.Buffer
			printApacheInstructions(&output, tc.host, certPath, keyPath, caPath)
			for _, want := range []string{
				certPath + " -> server.pem", keyPath + " -> server-key.pem", caPath,
				"-m 600 server-key.pem /etc/ssl/private/questlab-cert-key.pem",
				"/etc/apache2/sites-available/questlab-https.conf", "<VirtualHost *:443>",
				"ServerName " + tc.serverName, "DocumentRoot /var/www/html", "SSLEngine on",
				"SSLCertificateFile /etc/ssl/certs/questlab-cert.pem",
				"SSLCertificateKeyFile /etc/ssl/private/questlab-cert-key.pem",
				"sudo a2enmod ssl", "sudo a2ensite questlab-https.conf",
				"sudo apache2ctl configtest && sudo systemctl reload apache2",
				"/usr/local/share/ca-certificates/questlab-rootCA.crt", "sudo update-ca-certificates",
				"https://" + tc.serverName + "/", tc.warning, tc.caWarning,
			} {
				if !strings.Contains(output.String(), want) {
					t.Errorf("Apache instructions omit %q: %s", want, &output)
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
