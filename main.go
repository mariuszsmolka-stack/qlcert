// Copyright 2018 The mkcert Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// QuestLab Cert is based on mkcert, a tool for local development certificates.
package main

import (
	"crypto"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"

	"golang.org/x/net/idna"
)

// Version can be set at link time to override debug.BuildInfo.Main.Version,
// which is "(devel)" when building from within the module. See
// golang.org/issue/29814 and golang.org/issue/29228.
var Version string

func main() {
	uiLanguage = "en"
	if len(os.Args) == 1 {
		fmt.Print(helpText(uiLanguage, false))
		return
	}
	log.SetFlags(0)
	var (
		languageFlag  = flag.String("lang", "en", "")
		installFlag   = flag.Bool("install", false, "")
		uninstallFlag = flag.Bool("uninstall", false, "")
		pkcs12Flag    = flag.Bool("pkcs12", false, "")
		ecdsaFlag     = flag.Bool("ecdsa", false, "")
		clientFlag    = flag.Bool("client", false, "")
		helpFlag      = flag.Bool("help", false, "")
		carootFlag    = flag.Bool("CAROOT", false, "")
		csrFlag       = flag.String("csr", "", "")
		certFileFlag  = flag.String("cert-file", "", "")
		keyFileFlag   = flag.String("key-file", "", "")
		p12FileFlag   = flag.String("p12-file", "", "")
		versionFlag   = flag.Bool("version", false, "")
	)
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), helpText(*languageFlag, false))
	}
	flag.CommandLine.Usage = flag.Usage
	flag.CommandLine.SetOutput(languageFlagWriter{flag.CommandLine.Output(), languageFlag})
	flag.Parse()
	uiLanguage = *languageFlag
	if uiLanguage != "en" && uiLanguage != "pl" {
		log.Fatalf("ERROR: unsupported language %q; use -lang=en or -lang=pl.", uiLanguage)
	}
	if *helpFlag {
		fmt.Print(helpText(uiLanguage, true))
		return
	}
	if *versionFlag {
		version := Version
		if version == "" {
			if buildInfo, ok := debug.ReadBuildInfo(); ok {
				version = buildInfo.Main.Version
			} else {
				version = "(unknown)"
			}
		}
		fmt.Printf("QuestLab Cert %s\n", version)
		fmt.Println(tr("Based on mkcert by The mkcert Authors; BSD-3-Clause license (LICENSE, AUTHORS)."))
		return
	}
	if *carootFlag {
		if *installFlag || *uninstallFlag {
			log.Fatalln(tr("ERROR: you can't set -[un]install and -CAROOT at the same time"))
		}
		fmt.Println(getCAROOT())
		return
	}
	if *installFlag && *uninstallFlag {
		log.Fatalln(tr("ERROR: you can't set -install and -uninstall at the same time"))
	}
	if *csrFlag != "" && (*pkcs12Flag || *ecdsaFlag || *clientFlag) {
		log.Fatalln(tr("ERROR: can only combine -csr with -install and -cert-file"))
	}
	if *csrFlag != "" && flag.NArg() != 0 {
		log.Fatalln(tr("ERROR: can't specify extra arguments when using -csr"))
	}
	(&mkcert{
		installMode: *installFlag, uninstallMode: *uninstallFlag, csrPath: *csrFlag,
		pkcs12: *pkcs12Flag, ecdsa: *ecdsaFlag, client: *clientFlag,
		certFile: *certFileFlag, keyFile: *keyFileFlag, p12File: *p12FileFlag,
	}).Run(flag.Args())
}

const rootName = "rootCA.pem"
const rootKeyName = "rootCA-key.pem"

type mkcert struct {
	installMode, uninstallMode bool
	pkcs12, ecdsa, client      bool
	keyFile, certFile, p12File string
	csrPath                    string

	CAROOT string
	caCert *x509.Certificate
	caKey  crypto.PrivateKey

	// The system cert pool is only loaded once. After installing the root, checks
	// will keep failing until the next execution. TODO: maybe execve?
	// https://github.com/golang/go/issues/24540 (thanks, myself)
	ignoreCheckFailure bool
}

func (m *mkcert) Run(args []string) {
	if !m.uninstallMode {
		if len(args) == 0 && m.csrPath == "" && !m.installMode {
			flag.Usage()
			return
		}
		hostnameRegexp := regexp.MustCompile(`(?i)^(\*\.)?[0-9a-z_-]([0-9a-z._-]*[0-9a-z_-])?$`)
		for i, name := range args {
			if ip := net.ParseIP(name); ip != nil {
				continue
			}
			if email, err := mail.ParseAddress(name); err == nil && email.Address == name {
				continue
			}
			if uriName, err := url.Parse(name); err == nil && uriName.Scheme != "" && uriName.Host != "" {
				continue
			}
			punycode, err := idna.ToASCII(name)
			if err != nil {
				log.Fatalf(tr("ERROR: %q is not a valid hostname, IP, URL or email: %s"), name, err)
			}
			args[i] = punycode
			if !hostnameRegexp.MatchString(punycode) {
				log.Fatalf(tr("ERROR: %q is not a valid hostname, IP, URL or email"), name)
			}
		}
	}
	var csr *x509.CertificateRequest
	if m.csrPath != "" && !m.uninstallMode {
		csr = readCSR(m.csrPath)
	}

	m.CAROOT = getCAROOT()
	if m.CAROOT == "" {
		log.Fatalln(tr("ERROR: failed to find the default CA location, set one as the CAROOT env var"))
	}
	if !m.uninstallMode {
		fatalIfErr(os.MkdirAll(m.CAROOT, 0755), "failed to create the CAROOT")
		if !pathExists(filepath.Join(m.CAROOT, rootName)) {
			m.newCA()
		}
	}
	m.loadCA()

	if m.installMode {
		m.install()
		if len(args) == 0 {
			return
		}
	} else if m.uninstallMode {
		m.uninstall()
		return
	} else {
		var warning bool
		if storeEnabled("system") && !m.checkPlatform() {
			warning = true
			log.Println(tr("Note: the local CA is not installed in the system trust store."))
		}
		if storeEnabled("nss") && hasNSS && CertutilInstallHelp != "" && !m.checkNSS() {
			warning = true
			log.Printf(tr("Note: the local CA is not installed in the %s trust store."), tr(NSSBrowsers))
		}
		if storeEnabled("java") && hasJava && !m.checkJava() {
			warning = true
			log.Println(tr("Note: the local CA is not installed in the Java trust store."))
		}
		if warning {
			log.Println(tr("Run QuestLab Cert with -install for certificates to be trusted automatically ⚠️"))
		}
	}

	if m.csrPath != "" {
		m.makeCertFromCSR(csr)
		return
	}

	m.makeCert(args)
}

func getCAROOT() string {
	if env := os.Getenv("CAROOT"); env != "" {
		return env
	}

	var dir string
	switch {
	case runtime.GOOS == "windows":
		dir = os.Getenv("LocalAppData")
	case os.Getenv("XDG_DATA_HOME") != "":
		dir = os.Getenv("XDG_DATA_HOME")
	case runtime.GOOS == "darwin":
		dir = os.Getenv("HOME")
		if dir == "" {
			return ""
		}
		dir = filepath.Join(dir, "Library", "Application Support")
	default: // Unix
		dir = os.Getenv("HOME")
		if dir == "" {
			return ""
		}
		dir = filepath.Join(dir, ".local", "share")
	}
	return filepath.Join(dir, "mkcert")
}

func (m *mkcert) install() {
	if storeEnabled("system") {
		if m.checkPlatform() {
			log.Print(tr("The local CA is already installed in the system trust store! 👍"))
		} else {
			if m.installPlatform() {
				log.Print(tr("The local CA is now installed in the system trust store! ⚡️"))
			}
			m.ignoreCheckFailure = true // TODO: replace with a check for a successful install
		}
	}
	if storeEnabled("nss") && hasNSS {
		if m.checkNSS() {
			log.Printf(tr("The local CA is already installed in the %s trust store! 👍"), tr(NSSBrowsers))
		} else {
			if hasCertutil && m.installNSS() {
				log.Printf(tr("The local CA is now installed in the %s trust store (requires browser restart)! 🦊"), tr(NSSBrowsers))
			} else if CertutilInstallHelp == "" {
				log.Printf(tr(`Note: %s support is not available on your platform. ℹ️`), tr(NSSBrowsers))
			} else if !hasCertutil {
				log.Printf(tr(`Warning: "certutil" is not available, so the CA can't be automatically installed in %s! ⚠️`), tr(NSSBrowsers))
				log.Printf(tr(`Install "certutil" with "%s" and re-run QuestLab Cert with -install 👈`), CertutilInstallHelp)
			}
		}
	}
	if storeEnabled("java") && hasJava {
		if m.checkJava() {
			log.Println(tr("The local CA is already installed in Java's trust store! 👍"))
		} else {
			if hasKeytool {
				m.installJava()
				log.Println(tr("The local CA is now installed in Java's trust store! ☕️"))
			} else {
				log.Println(tr(`Warning: "keytool" is not available, so the CA can't be automatically installed in Java's trust store! ⚠️`))
			}
		}
	}
	log.Print("")
}

func (m *mkcert) uninstall() {
	if storeEnabled("nss") && hasNSS {
		if hasCertutil {
			m.uninstallNSS()
		} else if CertutilInstallHelp != "" {
			log.Print("")
			log.Printf(tr(`Warning: "certutil" is not available, so the CA can't be automatically uninstalled from %s (if it was ever installed)! ⚠️`), tr(NSSBrowsers))
			log.Printf(tr(`You can install "certutil" with "%s" and re-run QuestLab Cert with -uninstall 👈`), CertutilInstallHelp)
			log.Print("")
		}
	}
	if storeEnabled("java") && hasJava {
		if hasKeytool {
			m.uninstallJava()
		} else {
			log.Print("")
			log.Println(tr(`Warning: "keytool" is not available, so the CA can't be automatically uninstalled from Java's trust store (if it was ever installed)! ⚠️`))
			log.Print("")
		}
	}
	if storeEnabled("system") && m.uninstallPlatform() {
		log.Print(tr("The local CA is now uninstalled from the system trust store(s)! 👋"))
		log.Print("")
	} else if storeEnabled("nss") && hasCertutil {
		log.Printf(tr("The local CA is now uninstalled from the %s trust store(s)! 👋"), tr(NSSBrowsers))
		log.Print("")
	}
}

func (m *mkcert) checkPlatform() bool {
	if m.ignoreCheckFailure {
		return true
	}

	_, err := m.caCert.Verify(x509.VerifyOptions{})
	return err == nil
}

func storeEnabled(name string) bool {
	stores := os.Getenv("TRUST_STORES")
	if stores == "" {
		return true
	}
	for _, store := range strings.Split(stores, ",") {
		if store == name {
			return true
		}
	}
	return false
}

func fatalIfErr(err error, msg string) {
	if err != nil {
		log.Fatalf(tr("ERROR: %s: %s"), tr(msg), err)
	}
}

func fatalIfCmdErr(err error, cmd string, out []byte) {
	if err != nil {
		log.Fatalf(tr("ERROR: failed to execute \"%s\": %s\n\n%s\n"), cmd, err, out)
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func binaryExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

var sudoWarningOnce sync.Once

func commandWithSudo(cmd ...string) *exec.Cmd {
	if u, err := user.Current(); err == nil && u.Uid == "0" {
		return exec.Command(cmd[0], cmd[1:]...)
	}
	if !binaryExists("sudo") {
		sudoWarningOnce.Do(func() {
			log.Println(tr(`Warning: "sudo" is not available, and QuestLab Cert is not running as root. The (un)install operation might fail. ⚠️`))
		})
		return exec.Command(cmd[0], cmd[1:]...)
	}
	return exec.Command("sudo", append([]string{"--prompt=" + tr("Sudo password:"), "--"}, cmd...)...)
}
