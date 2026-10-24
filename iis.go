package main

import (
	"fmt"
	"io"
	"net"
)

func printIISInstructions(output io.Writer, host, certPath, keyPath, caPath string) {
	fmt.Fprintln(output, tr("\nIIS HTTPS — Windows Server (manual steps)"))
	fmt.Fprintln(output, tr("1. Use the server DNS name or IP contained in the certificate. localhost refers to the client itself, not a remote server. If the name/IP is wrong, generate another certificate. Configure client DNS or its hosts file to resolve the server name."))
	fmt.Fprintln(output, tr("2. IIS conveniently imports a certificate AND its server private key as a password-protected PFX (PKCS#12) file. PEM alone is not that bundle. These PEM files are inputs to separate PFX preparation:"))
	fmt.Fprintf(output, "  %s\n  %s\n", certPath, keyPath)
	fmt.Fprintln(output, tr("LIMITATION B01/B02: QuestLab Cert currently exports PFX with the fixed password changeit and legacy algorithms. Do NOT use that export as a secure solution. Full integration requires fixing B01/B02. For now, ask an administrator to prepare a PFX using a separately verified tool with strong protection and a unique password. Include only the server key, never rootCA-key.pem."))
	fmt.Fprintln(output, tr("3. Continue ONLY with that separately prepared PFX. Securely transfer it to C:\\Certificates\\server.pfx on the IIS server. In PowerShell as Administrator, import into Local Computer > Personal (My); enter the password at the hidden prompt:"))
	fmt.Fprintln(output, `Import-PfxCertificate -FilePath 'C:\Certificates\server.pfx' -CertStoreLocation Cert:\LocalMachine\My -Password (Read-Host -Prompt 'PFX password' -AsSecureString)`)
	fmt.Fprintln(output, tr("Run certlm.msc > Personal > Certificates. Open the imported server certificate and confirm that it has a private key. Keep the PFX and server key private, outside the website directory."))
	fmt.Fprintln(output, tr("4. Run inetmgr (IIS Manager) > Sites > your existing HTTP site > Bindings > Add. Set Type=https, Port=443, select the server IP (or All Unassigned) and the imported certificate under SSL certificate."))
	if net.ParseIP(host) == nil {
		fmt.Fprintf(output, tr("Set Host name=%s. When sharing an IP and port 443 between DNS sites, select Require Server Name Indication (SNI) for those bindings.\n"), host)
	} else {
		fmt.Fprintf(output, tr("For access by IP %s, leave Host name empty and SNI unchecked; use a dedicated IP:443 binding with this certificate.\n"), host)
	}
	fmt.Fprintln(output, tr("Confirm the binding with OK. Do not replace another site's binding on the same IP:443. Importing into Personal does not establish client trust; follow the client steps below."))
	printClientTrustInstructions(output, host, caPath)
}
