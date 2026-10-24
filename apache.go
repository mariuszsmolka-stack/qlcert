package main

import (
	"fmt"
	"io"
	"strings"
)

// Display manual setup only; never run commands or write configuration files.
func printApacheInstructions(output io.Writer, host, certPath, keyPath, caPath string) {
	serverName := host
	if strings.Contains(host, ":") {
		serverName = "[" + host + "]" // IPv6 literals in ServerName and HTTPS URLs.
	}
	fmt.Fprintln(output, tr("\nApache 2.4 HTTPS — Ubuntu/Debian (manual steps)"))
	fmt.Fprintln(output, tr("1. Use the server DNS name or IP contained in the certificate. localhost refers to the client itself, not a remote server. If the name/IP is wrong, generate another certificate. Configure client DNS or its hosts file to resolve the server name."))
	fmt.Fprintln(output, tr("2. Securely transfer these files to the Ubuntu/Debian server, naming them server.pem and server-key.pem in your working directory:"))
	fmt.Fprintf(output, "  %s -> server.pem\n  %s -> server-key.pem\n", certPath, keyPath)
	fmt.Fprintln(output, tr("Keep the server private key secret and outside the website directory. On the server, run:"))
	fmt.Fprintln(output, `sudo install -o root -g root -m 644 server.pem /etc/ssl/certs/questlab-cert.pem
sudo install -o root -g root -m 600 server-key.pem /etc/ssl/private/questlab-cert-key.pem`)
	fmt.Fprintln(output, tr("3. Create /etc/apache2/sites-available/questlab-https.conf. Use the DocumentRoot and application directives from your working HTTP site; /var/www/html below is an example:"))
	fmt.Fprintf(output, `<VirtualHost *:443>
    ServerName %s
    DocumentRoot /var/www/html
    SSLEngine on
    SSLCertificateFile /etc/ssl/certs/questlab-cert.pem
    SSLCertificateKeyFile /etc/ssl/private/questlab-cert-key.pem
</VirtualHost>
`, serverName)
	fmt.Fprintln(output, tr("4. Enable SSL and the site. Ensure /etc/apache2/ports.conf listens on 443 (the standard SSL configuration does); do not add a duplicate Listen directive. Reload only if configtest reports Syntax OK:"))
	fmt.Fprintln(output, `sudo a2enmod ssl
sudo a2ensite questlab-https.conf
sudo apache2ctl configtest && sudo systemctl reload apache2`)
	printClientTrustInstructions(output, host, caPath)
}

func printClientTrustInstructions(output io.Writer, host, caPath string) {
	serverName := host
	if strings.Contains(host, ":") {
		serverName = "[" + host + "]"
	}
	fmt.Fprintln(output, tr("Generating a certificate does not make browsers trust it. The client must trust the local CA (the authority that signed it)."))
	fmt.Fprintln(output, tr("5. On EACH client, trust the public CA certificate from:"))
	fmt.Fprintln(output, "  "+caPath)
	fmt.Fprintln(output, tr("Copy ONLY rootCA.pem from this CA to the client. NEVER copy or share rootCA-key.pem (the private CA key). On an Ubuntu/Debian client with ca-certificates installed, from the directory containing rootCA.pem run:"))
	fmt.Fprintln(output, `sudo install -o root -g root -m 644 rootCA.pem /usr/local/share/ca-certificates/questlab-rootCA.crt
sudo update-ca-certificates`)
	fmt.Fprintln(output, tr("On Windows, run certmgr.msc: Trusted Root Certification Authorities > Certificates > All Tasks > Import; select rootCA.pem (All Files) for the current user. If the browser uses a separate trust store (e.g. Firefox), import the public CA there as well and allow it to identify websites. Restart the browser."))
	fmt.Fprintf(output, tr("Open %s on the client; check that the page loads without a certificate warning. The server must be reachable on TCP 443. QuestLab Cert only displays these steps; it does not configure services or install trust."), "https://"+serverName+"/")
	fmt.Fprintln(output)
}
