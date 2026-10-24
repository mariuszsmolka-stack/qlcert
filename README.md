# QuestLab Cert

QuestLab Cert helps students, teachers and administrators set up HTTPS in local labs. It generates certificates signed by a local certificate authority (CA) and provides practical server setup instructions.

The project is based on [mkcert](https://github.com/FiloSottile/mkcert) by The mkcert Authors. Mariusz Smółka / [QuestLab](https://questlab.pl) develops and adapts it as QuestLab Cert. Original authorship is preserved in [AUTHORS](AUTHORS), and the project retains the [BSD-3-Clause license](LICENSE).

## Current features

- English by default; Polish with `-lang=pl`, or explicit English with `-lang=en`.
- An interactive wizard for one DNS name or IP address, with PEM for Apache/Nginx or PFX for IIS, a summary and confirmation before generation.
- PFX export using the explicit **Modern2023** profile. The password and confirmation are entered without echo in an interactive terminal; passwords are not accepted through CLI arguments, environment variables or pipes, and are not logged.
- Protection against overwriting existing PEM/PFX output files, including restricted and verified PFX permissions on Windows.
- Manual HTTPS instructions for Apache 2.4 and Nginx on Ubuntu/Debian, and IIS on Windows Server.
- The traditional CLI, including certificates with multiple DNS names and IP addresses.

Wizard output goes to `QuestLab-Cert/certificates` inside your home directory. The wizard does not automatically install CA trust or configure web servers.

## Quick start

Build from source with Go 1.26 or newer. On Windows:

```powershell
go build -o questlab-cert.exe .

# Open the English wizard
.\questlab-cert.exe

# Open the Polish wizard
.\questlab-cert.exe -lang=pl

# Show full help
.\questlab-cert.exe -help
```

Traditional CLI examples:

```powershell
# PEM certificate for a server DNS name and IP address
.\questlab-cert.exe server.lab 192.168.1.20

# PFX export; the terminal asks for a password and confirmation
.\questlab-cert.exe -pkcs12 -p12-file server.pfx server.lab

# Show the CA directory
.\questlab-cert.exe -CAROOT
```

CLI output defaults to the working directory; output path options remain available. On Linux, build with `go build -o questlab-cert .` and run `./questlab-cert` with the same options.

Replace the example names and IP with your server's actual address. Clients must resolve the DNS name to that server, and the certificate must cover the name or IP used to connect. `localhost` refers to the client's own computer.

## CA trust and private keys

**Never share `rootCA-key.pem` or copy it to a web server or client.** This private key can sign certificates on behalf of your lab CA. Keep server private keys and PFX files protected too, and choose a long, unique random PFX password. PFX exports include the server key and the public CA certificate, never the private CA key.

Generating a certificate does not make a browser trust it. Distribute only the public `rootCA.pem` to clients that should trust your lab CA, then follow the displayed trust instructions. Trust installation is a separate, explicit action; the traditional CLI also retains `-install`.

## Development status

QuestLab Cert is under active development for laboratory use. Automated tests and interactive checks have run on Windows; Linux has been checked by cross-compilation. Tests on real Apache, Nginx and IIS servers are not complete. PFX import into a real IIS instance and runtime behavior on Linux still require verification.

See the [roadmap](ROADMAP.md) and [development history](docs/DEVELOPMENT_HISTORY.md) for completed work, open tasks and test limits. The [original mkcert project](https://github.com/FiloSottile/mkcert) contains its upstream documentation.
