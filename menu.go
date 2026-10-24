package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/idna"
)

func runMenu(input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(menuReader{input})
	for {
		fmt.Fprintln(output, "\nQuestLab Cert")
		fmt.Fprintln(output, tr("1. Generate HTTPS certificate"))
		fmt.Fprintln(output, tr("2. Help"))
		fmt.Fprintln(output, tr("0. Exit"))
		fmt.Fprint(output, tr("Choose an option: "))
		if !scanner.Scan() {
			return scanner.Err()
		}
		switch strings.TrimSpace(scanner.Text()) {
		case "0", "3":
			return nil
		case "2":
			fmt.Fprint(output, helpText(uiLanguage, true))
		case "1":
			fmt.Fprint(output, tr("Enter one DNS name or IP address (empty to cancel): "))
			if !scanner.Scan() {
				return scanner.Err()
			}
			name := strings.TrimSpace(scanner.Text())
			if name == "" {
				fmt.Fprintln(output, tr("Cancelled. No files were created."))
				continue
			}
			host, valid := menuHost(name)
			if !valid {
				fmt.Fprintln(output, tr("Invalid input. Enter a single DNS name or IP address."))
				continue
			}
			var format string
			for {
				fmt.Fprintln(output, tr("1. PEM - Apache / Nginx (default)"))
				fmt.Fprintln(output, "2. PFX - IIS")
				fmt.Fprintln(output, tr("0. Back"))
				fmt.Fprint(output, tr("Choose a format [1]: "))
				if !scanner.Scan() {
					return scanner.Err()
				}
				format = strings.TrimSpace(scanner.Text())
				if format == "" || format == "1" || format == "2" || format == "0" {
					break
				}
				fmt.Fprintln(output, tr("Invalid format. Choose 1, 2 or 0."))
			}
			if format == "0" {
				continue
			}
			m := &mkcert{pkcs12: format == "2"} // Never install trust from the menu.
			certPath, keyPath, p12Path := m.fileNames([]string{host})
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf(tr("failed to locate the home directory: %w"), err)
			}
			certDir, err := filepath.Abs(filepath.Join(home, "QuestLab-Cert", "certificates"))
			if err != nil {
				return fmt.Errorf(tr("failed to resolve the certificate directory: %w"), err)
			}
			m.certFile = filepath.Join(certDir, certPath)
			m.keyFile = filepath.Join(certDir, keyPath)
			m.p12File = filepath.Join(certDir, strings.TrimSuffix(filepath.Base(p12Path), ".p12")+".pfx")
			if m.pkcs12 {
				fmt.Fprintf(output, tr("\nSummary\nDNS/IP: %s\nBundle: %s\nFormat: PFX (Modern2023)\n"), host, m.p12File)
			} else {
				fmt.Fprintf(output, tr("\nSummary\nDNS/IP: %s\nCertificate: %s\nPrivate key: %s\nFormat: PEM\n"), host, m.certFile, m.keyFile)
			}
			fmt.Fprintln(output, tr("A local CA will be created if needed. Trust will not be installed."))
			fmt.Fprint(output, tr("Generate the certificate? [y/N]: "))
			if !scanner.Scan() {
				return scanner.Err()
			}
			answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if answer != "y" && answer != "yes" && answer != "t" && answer != "tak" {
				fmt.Fprintln(output, tr("Cancelled. No files were created."))
				continue
			}
			m.run([]string{host}, certDir)
			if m.pkcs12 {
				fmt.Fprint(output, tr("Show IIS HTTPS instructions? [y/N]: "))
				if !scanner.Scan() {
					return scanner.Err()
				}
				switch strings.ToLower(strings.TrimSpace(scanner.Text())) {
				case "y", "yes", "t", "tak":
					printIISInstructions(output, host, "", "", filepath.Join(m.CAROOT, rootName), m.p12File)
				}
			} else if err := showHTTPSInstructions(scanner, output, host, m.certFile, m.keyFile, filepath.Join(m.CAROOT, rootName)); err != nil {
				return err
			}
		default:
			fmt.Fprintln(output, tr("Invalid option. Choose 1, 2 or 0."))
		}
	}
}

// Scanner must not read ahead into the password later read from the terminal.
type menuReader struct{ io.Reader }

func (r menuReader) Read(p []byte) (int, error) {
	if len(p) > 1 {
		p = p[:1]
	}
	return r.Reader.Read(p)
}

func showHTTPSInstructions(scanner *bufio.Scanner, output io.Writer, host, certPath, keyPath, caPath string) error {
	for {
		fmt.Fprintln(output, tr("\nHTTPS setup instructions"))
		fmt.Fprintln(output, "1. Apache\n2. Nginx\n3. IIS")
		fmt.Fprintln(output, tr("4. Back"))
		fmt.Fprint(output, tr("Choose an option: "))
		if !scanner.Scan() {
			return scanner.Err()
		}
		switch strings.TrimSpace(scanner.Text()) {
		case "1":
			printApacheInstructions(output, host, certPath, keyPath, caPath)
		case "2":
			printNginxInstructions(output, host, certPath, keyPath, caPath)
		case "3":
			printIISInstructions(output, host, certPath, keyPath, caPath, "")
		case "4":
			return nil
		default:
			fmt.Fprintln(output, tr("Invalid option. Choose 1, 2, 3 or 4."))
		}
	}
}

// The menu accepts one DNS hostname or IP, not CLI options, URLs or email SANs.
// Leave the existing, broader CLI validation and certificate generation intact.
func menuHost(name string) (string, bool) {
	if net.ParseIP(name) != nil {
		return name, true
	}
	host, err := idna.ToASCII(name)
	if err != nil {
		return "", false
	}
	dns := strings.TrimSuffix(host, ".")
	if len(dns) == 0 || len(dns) > 253 {
		return "", false
	}
	if strings.Contains(dns, ".") && strings.Trim(dns, "0123456789.") == "" {
		return "", false // An invalid dotted IP must not be accepted as DNS.
	}
	for _, label := range strings.Split(dns, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", false
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-') {
				return "", false
			}
		}
	}
	return host, true
}
