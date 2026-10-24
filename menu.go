package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"

	"golang.org/x/net/idna"
)

func runMenu(input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)
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
			m := &mkcert{} // PEM only; never enable installation of trust here.
			certPath, keyPath, _ := m.fileNames([]string{host})
			fmt.Fprintf(output, tr("\nSummary\nDNS/IP: %s\nCertificate: %s\nPrivate key: %s\nFormat: PEM\n"), host, certPath, keyPath)
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
			m.Run([]string{host})
		default:
			fmt.Fprintln(output, tr("Invalid option. Choose 1, 2 or 0."))
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
