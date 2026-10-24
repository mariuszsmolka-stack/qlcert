package main

import (
	"fmt"
	"io"
)

func printNginxInstructions(output io.Writer, host, certPath, keyPath, caPath string) {
	fmt.Fprintln(output, tr("\nNginx HTTPS — Ubuntu/Debian (manual steps)"))
	fmt.Fprintln(output, tr("1. Use the server DNS name or IP contained in the certificate. localhost refers to the client itself, not a remote server. If the name/IP is wrong, generate another certificate. Configure client DNS or its hosts file to resolve the server name."))
	fmt.Fprintln(output, tr("2. Securely transfer these files to the Ubuntu/Debian server, naming them server.pem and server-key.pem in your working directory:"))
	fmt.Fprintf(output, "  %s -> server.pem\n  %s -> server-key.pem\n", certPath, keyPath)
	fmt.Fprintln(output, tr("Keep the server private key secret and outside the website directory. On the server, run:"))
	fmt.Fprintln(output, `sudo install -o root -g root -m 644 server.pem /etc/ssl/certs/questlab-cert.pem
sudo install -o root -g root -m 600 server-key.pem /etc/ssl/private/questlab-cert-key.pem`)
	fmt.Fprintln(output, tr("3. Create /etc/nginx/conf.d/questlab-https.conf. Check that /etc/nginx/nginx.conf includes conf.d/*.conf inside http. Use root and application settings from your working HTTP site; /var/www/html is an example:"))
	fmt.Fprintf(output, `server {
    listen 443 ssl;
    listen [::]:443 ssl;
    server_name %s;
    root /var/www/html;
    index index.html;
    ssl_certificate /etc/ssl/certs/questlab-cert.pem;
    ssl_certificate_key /etc/ssl/private/questlab-cert-key.pem;
}
`, host)
	fmt.Fprintln(output, tr("4. Check the configuration; reload only if the check succeeds:"))
	fmt.Fprintln(output, "sudo nginx -t && sudo systemctl reload nginx")
	printClientTrustInstructions(output, host, caPath)
}
