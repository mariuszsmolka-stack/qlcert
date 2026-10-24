package main

const helpShortEN = `QuestLab Cert — local HTTPS certificates
Based on mkcert by The mkcert Authors; BSD-3-Clause license (LICENSE, AUTHORS).

Usage: PROGRAM [options] NAME [NAME ...]
PROGRAM means the name or path of your QuestLab Cert executable.
Without arguments, or with -lang only, the program opens the console menu.

  PROGRAM -install
      Create or load the local CA and install its certificate as trusted.
  PROGRAM example.test localhost 127.0.0.1 ::1
      Generate a certificate and private key for these names and addresses.
  PROGRAM "*.example.test"
      Generate a wildcard certificate (one DNS level).
  PROGRAM -uninstall
      Remove trust in the existing CA; keep its files.
  PROGRAM -lang=pl -help
      Show the full help in Polish. The default language is English.

Use -help for the full list of options. Put options before the list of names.
`

const helpShortPL = `QuestLab Cert — lokalne certyfikaty HTTPS
Na podstawie mkcert autorstwa The mkcert Authors; licencja BSD-3-Clause (LICENSE, AUTHORS).

Użycie: PROGRAM [opcje] NAZWA [NAZWA ...]
PROGRAM oznacza nazwę lub ścieżkę pliku wykonywalnego QuestLab Cert.
Bez argumentów lub z samym -lang program otwiera menu konsolowe.

  PROGRAM -install
      Utwórz lub odczytaj lokalne CA i zainstaluj zaufanie do jego certyfikatu.
  PROGRAM example.test localhost 127.0.0.1 ::1
      Wygeneruj certyfikat i klucz prywatny dla tych nazw i adresów.
  PROGRAM "*.example.test"
      Wygeneruj certyfikat wildcard (jeden poziom DNS).
  PROGRAM -uninstall
      Usuń zaufanie do istniejącego CA; zachowaj jego pliki.
  PROGRAM -lang=en -help
      Wyświetl pełną pomoc po angielsku. Domyślnym językiem jest angielski.

Użyj -help, aby wyświetlić wszystkie opcje. Opcje podawaj przed listą nazw.
`

const helpOptionsEN = `
Options:
  -help
      Show this full help and exit. -h shows the short usage.
  -version
      Show the program name, version and mkcert attribution, then exit.
  -lang=en, -lang=pl
      Select English or Polish for this invocation only (default: en).
      No system-language detection and no saved language preferences.
  -install
      Install trust in the local CA in the selected trust stores.
      May be combined with certificate generation.
  -uninstall
      Remove trust in the existing local CA without deleting its files.
      Cannot be combined with -install or -CAROOT.
  -CAROOT
      Print the CA directory without creating a CA; cannot be combined with
      -install or -uninstall.
  -cert-file FILE, -key-file FILE, -p12-file FILE
      Set the certificate, private-key or PKCS#12 output path.
      Equal -cert-file and -key-file paths produce a combined PEM file.
      Existing PEM output files are not overwritten.
  -client
      Include client authentication among the certificate's uses.
  -ecdsa
      Use an ECDSA key instead of the default RSA key.
  -pkcs12
      Export a Modern2023 PKCS#12 (.p12/.pfx) bundle containing the certificate,
      its private key and the CA certificate instead of separate PEM files.
      Requires a password entered twice without echo in an interactive terminal.
      Existing PFX files are not overwritten. Never pass a password as an argument.
  -csr FILE
      Sign a PEM certificate signing request. Cannot be combined with names,
      -pkcs12, -ecdsa or -client. Supports -install and -cert-file.

Names may be DNS names (including international names), IPv4/IPv6 addresses,
absolute URIs or email addresses. Email addresses also enable S/MIME usage.

Environment variables:
  CAROOT
      Override the existing CA storage location; allows separate local CAs.
  TRUST_STORES
      Comma-separated trust stores: system, java, nss (includes Firefox).
      When unset or empty, available stores are selected automatically.
  JAVA_HOME
      Locate the Java installation and its trust store (keytool/cacerts).

All existing option names remain the same in both languages. Options must
precede names. Generation does not install trust without -install. QuestLab
Cert does not configure the web server. Never share rootCA-key.pem.
`

const helpOptionsPL = `
Opcje:
  -help
      Wyświetl tę pełną pomoc i zakończ. -h wyświetla skróconą pomoc.
  -version
      Wyświetl nazwę, wersję i informację o pochodzeniu z mkcert, a następnie zakończ.
  -lang=en, -lang=pl
      Wybierz angielski lub polski tylko dla bieżącego uruchomienia (domyślnie: en).
      Bez wykrywania języka systemu i zapisywania preferencji językowych.
  -install
      Zainstaluj zaufanie do lokalnego CA w wybranych magazynach.
      Można połączyć z generowaniem certyfikatu.
  -uninstall
      Usuń zaufanie do istniejącego lokalnego CA, zachowując jego pliki.
      Nie można łączyć z -install ani -CAROOT.
  -CAROOT
      Wyświetl katalog CA bez tworzenia CA; nie można łączyć z -install
      ani -uninstall.
  -cert-file PLIK, -key-file PLIK, -p12-file PLIK
      Ustaw ścieżkę certyfikatu, klucza prywatnego lub pakietu PKCS#12.
      Jednakowe ścieżki -cert-file i -key-file tworzą wspólny plik PEM.
      Istniejące pliki wynikowe PEM nie są nadpisywane.
  -client
      Dodaj uwierzytelnianie klienta do zastosowań certyfikatu.
  -ecdsa
      Użyj klucza ECDSA zamiast domyślnego klucza RSA.
  -pkcs12
      Eksportuj PKCS#12 Modern2023 (.p12/.pfx) z certyfikatem, jego kluczem
      prywatnym i certyfikatem CA zamiast osobnych plików PEM.
      Wymaga dwukrotnego podania hasła bez echa w interaktywnym terminalu.
      Nie nadpisuje PFX. Nigdy nie podawaj hasła jako argumentu.
  -csr PLIK
      Podpisz żądanie certyfikatu w formacie PEM. Nie można łączyć z nazwami,
      -pkcs12, -ecdsa ani -client. Obsługuje -install i -cert-file.

Nazwy mogą być nazwami DNS (także międzynarodowymi), adresami IPv4/IPv6,
bezwzględnymi URI lub adresami e-mail. Adres e-mail dodaje zastosowanie S/MIME.

Zmienne środowiskowe:
  CAROOT
      Wskaż lokalizację przechowywania CA; umożliwia rozdzielenie lokalnych CA.
  TRUST_STORES
      Magazyny oddzielone przecinkami: system, java, nss (obejmuje Firefox).
      Brak lub pusta wartość oznacza automatyczny wybór dostępnych magazynów.
  JAVA_HOME
      Wskaż instalację Java i jej magazyn zaufania (keytool/cacerts).

Nazwy wszystkich istniejących parametrów są takie same w obu językach.
Opcje muszą poprzedzać nazwy. Generowanie nie instaluje zaufania bez -install.
QuestLab Cert nie konfiguruje serwera WWW. Nigdy nie udostępniaj rootCA-key.pem.
`

func helpText(language string, full bool) string {
	short, options := helpShortEN, helpOptionsEN
	if language == "pl" {
		short, options = helpShortPL, helpOptionsPL
	}
	if full {
		return short + options
	}
	return short
}
