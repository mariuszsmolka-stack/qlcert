package main

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// Set once by main for this invocation; no environment detection or persistence.
var uiLanguage = "en"

func tr(message string) string {
	if uiLanguage == "pl" {
		if translated, ok := polishMessages[message]; ok {
			return translated
		}
	}
	return message
}

// System/library errors and subprocess output remain in their original language.
var polishMessages = map[string]string{
	"1. Generate HTTPS certificate": "1. Utwórz certyfikat HTTPS",
	"2. Help":                       "2. Pomoc",
	"0. Exit":                       "0. Zakończ",
	"Choose an option: ":            "Wybierz opcję: ",
	"Enter one DNS name or IP address (empty to cancel): ":                            "Podaj jedną nazwę DNS lub adres IP (puste pole anuluje): ",
	"Cancelled. No files were created.":                                               "Anulowano. Nie utworzono żadnych plików.",
	"Invalid input. Enter a single DNS name or IP address.":                           "Nieprawidłowe dane. Podaj jedną nazwę DNS lub adres IP.",
	"\nSummary\nDNS/IP: %s\nCertificate: %s\nPrivate key: %s\nFormat: PEM\n":          "\nPodsumowanie\nDNS/IP: %s\nCertyfikat: %s\nKlucz prywatny: %s\nFormat: PEM\n",
	"A local CA will be created if needed. Trust will not be installed.":              "W razie potrzeby powstanie lokalne CA. Zaufanie nie zostanie zainstalowane.",
	"Generate the certificate? [y/N]: ":                                               "Utworzyć certyfikat? [t/N]: ",
	"Invalid option. Choose 1, 2 or 0.":                                               "Nieprawidłowa opcja. Wybierz 1, 2 lub 0.",
	"failed to read menu input":                                                       "nie udało się odczytać danych wejściowych menu",
	"Based on mkcert by The mkcert Authors; BSD-3-Clause license (LICENSE, AUTHORS).": "Na podstawie mkcert autorstwa The mkcert Authors; licencja BSD-3-Clause (LICENSE, AUTHORS).",
	"ERROR: %s: %s": "BŁĄD: %s: %s",
	"ERROR: failed to execute \"%s\": %s\n\n%s\n":                                                                                              "BŁĄD: nie udało się wykonać \"%s\": %s\n\n%s\n",
	"ERROR: you can't set -[un]install and -CAROOT at the same time":                                                                           "BŁĄD: nie można jednocześnie używać -[un]install i -CAROOT",
	"ERROR: you can't set -install and -uninstall at the same time":                                                                            "BŁĄD: nie można jednocześnie używać -install i -uninstall",
	"ERROR: can only combine -csr with -install and -cert-file":                                                                                "BŁĄD: -csr można łączyć tylko z -install i -cert-file",
	"ERROR: can't specify extra arguments when using -csr":                                                                                     "BŁĄD: nie można podawać dodatkowych argumentów przy użyciu -csr",
	"ERROR: %q is not a valid hostname, IP, URL or email: %s":                                                                                  "BŁĄD: %q nie jest poprawną nazwą hosta, adresem IP, URL ani e-mail: %s",
	"ERROR: %q is not a valid hostname, IP, URL or email":                                                                                      "BŁĄD: %q nie jest poprawną nazwą hosta, adresem IP, URL ani e-mail",
	"ERROR: failed to find the default CA location, set one as the CAROOT env var":                                                             "BŁĄD: nie ustalono domyślnej lokalizacji CA; ustaw zmienną środowiskową CAROOT",
	"Note: the local CA is not installed in the system trust store.":                                                                           "Uwaga: lokalne CA nie jest zainstalowane w systemowym magazynie zaufania.",
	"Note: the local CA is not installed in the %s trust store.":                                                                               "Uwaga: lokalne CA nie jest zainstalowane w magazynie zaufania %s.",
	"Note: the local CA is not installed in the Java trust store.":                                                                             "Uwaga: lokalne CA nie jest zainstalowane w magazynie zaufania Java.",
	"Run QuestLab Cert with -install for certificates to be trusted automatically ⚠️":                                                          "Uruchom QuestLab Cert z -install, aby zainstalować zaufanie do certyfikatów ⚠️",
	"The local CA is already installed in the system trust store! 👍":                                                                           "Lokalne CA jest już zainstalowane w systemowym magazynie zaufania! 👍",
	"The local CA is now installed in the system trust store! ⚡️":                                                                              "Lokalne CA zostało zainstalowane w systemowym magazynie zaufania! ⚡️",
	"The local CA is already installed in the %s trust store! 👍":                                                                               "Lokalne CA jest już zainstalowane w magazynie zaufania %s! 👍",
	"The local CA is now installed in the %s trust store (requires browser restart)! 🦊":                                                        "Lokalne CA zostało zainstalowane w magazynie zaufania %s (wymaga ponownego uruchomienia przeglądarki)! 🦊",
	"Note: %s support is not available on your platform. ℹ️":                                                                                   "Uwaga: obsługa %s nie jest dostępna na tej platformie. ℹ️",
	`Warning: "certutil" is not available, so the CA can't be automatically installed in %s! ⚠️`:                                               `Ostrzeżenie: brak "certutil" uniemożliwia automatyczną instalację CA w %s! ⚠️`,
	`Install "certutil" with "%s" and re-run QuestLab Cert with -install 👈`:                                                                    `Zainstaluj "certutil" poleceniem "%s" i uruchom ponownie QuestLab Cert z -install 👈`,
	"The local CA is already installed in Java's trust store! 👍":                                                                               "Lokalne CA jest już zainstalowane w magazynie zaufania Java! 👍",
	"The local CA is now installed in Java's trust store! ☕️":                                                                                  "Lokalne CA zostało zainstalowane w magazynie zaufania Java! ☕️",
	`Warning: "keytool" is not available, so the CA can't be automatically installed in Java's trust store! ⚠️`:                                `Ostrzeżenie: brak "keytool" uniemożliwia automatyczną instalację CA w magazynie zaufania Java! ⚠️`,
	`Warning: "certutil" is not available, so the CA can't be automatically uninstalled from %s (if it was ever installed)! ⚠️`:                `Ostrzeżenie: brak "certutil" uniemożliwia automatyczne usunięcie CA z %s (jeżeli zostało zainstalowane)! ⚠️`,
	`You can install "certutil" with "%s" and re-run QuestLab Cert with -uninstall 👈`:                                                          `Możesz zainstalować "certutil" poleceniem "%s" i uruchomić ponownie QuestLab Cert z -uninstall 👈`,
	`Warning: "keytool" is not available, so the CA can't be automatically uninstalled from Java's trust store (if it was ever installed)! ⚠️`: `Ostrzeżenie: brak "keytool" uniemożliwia automatyczne usunięcie CA z magazynu zaufania Java (jeżeli zostało zainstalowane)! ⚠️`,
	"The local CA is now uninstalled from the system trust store(s)! 👋":                                                                        "Lokalne CA zostało usunięte z systemowych magazynów zaufania! 👋",
	"The local CA is now uninstalled from the %s trust store(s)! 👋":                                                                            "Lokalne CA zostało usunięte z magazynów zaufania %s! 👋",
	`Warning: "sudo" is not available, and QuestLab Cert is not running as root. The (un)install operation might fail. ⚠️`:                     `Ostrzeżenie: brak "sudo", a QuestLab Cert nie działa jako root. Instalacja lub usuwanie zaufania może się nie powieść. ⚠️`,
	"Sudo password:": "Hasło sudo:",
	"ERROR: can't create new certificates because the CA key (rootCA-key.pem) is missing":         "BŁĄD: nie można tworzyć certyfikatów, ponieważ brakuje klucza CA (rootCA-key.pem)",
	"\nThe certificate and key are at \"%s\" ✅\n\n":                                               "\nCertyfikat i klucz zapisano w \"%s\" ✅\n\n",
	"\nThe certificate is at \"%s\" and the key at \"%s\" ✅\n\n":                                  "\nCertyfikat zapisano w \"%s\", a klucz w \"%s\" ✅\n\n",
	"\nThe PKCS#12 bundle is at \"%s\" ✅\n":                                                       "\nPakiet PKCS#12 zapisano w \"%s\" ✅\n",
	"\nThe legacy PKCS#12 encryption password is the often hardcoded default \"changeit\" ℹ️\n\n": "\nHasło szyfrowania starszego formatu PKCS#12 to często używana stała wartość \"changeit\" ℹ️\n\n",
	"It will expire on %s 🗓\n\n":                                                                  "Certyfikat wygaśnie %s 🗓\n\n",
	"\nCreated a new certificate valid for the following names 📜":                                 "\nUtworzono nowy certyfikat dla następujących nazw 📜",
	"   Warning: many browsers don't support second-level wildcards like %q ⚠️":                   "   Ostrzeżenie: wiele przeglądarek nie obsługuje wildcard na drugim poziomie, np. %q ⚠️",
	"\nReminder: X.509 wildcards only go one level deep, so this won't match a.b.%s ℹ️":           "\nPrzypomnienie: wildcard X.509 obejmuje tylko jeden poziom, więc nie pasuje do a.b.%s ℹ️",
	"ERROR: failed to read the CSR: unexpected content":                                           "BŁĄD: nie udało się odczytać CSR: nieoczekiwana zawartość",
	"ERROR: failed to read the CSR: expected CERTIFICATE REQUEST, got ":                           "BŁĄD: nie udało się odczytać CSR: oczekiwano CERTIFICATE REQUEST, otrzymano ",
	"\nThe certificate is at \"%s\" ✅\n\n":                                                        "\nCertyfikat zapisano w \"%s\" ✅\n\n",
	"ERROR: failed to read the CA certificate: unexpected content":                                "BŁĄD: nie udało się odczytać certyfikatu CA: nieoczekiwana zawartość",
	"ERROR: failed to read the CA key: unexpected content":                                        "BŁĄD: nie udało się odczytać klucza CA: nieoczekiwana zawartość",
	"Created a new local CA 💥\n":                                                                  "Utworzono nowe lokalne CA 💥\n",
	"Installing to the system store is not yet supported on this Linux 😣 but %s will still work.": "Instalacja w systemowym magazynie nie jest jeszcze obsługiwana w tej dystrybucji Linux 😣, ale %s nadal będzie działać.",
	"You can also manually install the root certificate at %q.":                                   "Możesz też ręcznie zainstalować certyfikat główny z %q.",
	"ERROR: no %s security databases found":                                                       "BŁĄD: nie znaleziono baz zabezpieczeń %s",
	"Installing in %s failed. Please report the issue with details about your environment at https://github.com/FiloSottile/mkcert/issues/new 👎": "Instalacja w %s nie powiodła się. Zgłoś problem ze szczegółami środowiska: https://github.com/FiloSottile/mkcert/issues/new 👎",
	"Note that if you never started %s, you need to do that at least once.":                                                                      "Jeśli %s nie było jeszcze uruchamiane, uruchom je przynajmniej raz.",
	"ERROR: unsupported trust settings version:":                                                                                                 "BŁĄD: nieobsługiwana wersja ustawień zaufania:",
	"Firefox and/or Chrome/Chromium":                                                                                                             "Firefox i/lub Chrome/Chromium",
	"failed to create the CAROOT":                                                                                                                "nie udało się utworzyć katalogu CAROOT",
	"failed to generate certificate key":                                                                                                         "nie udało się wygenerować klucza certyfikatu",
	"failed to generate certificate":                                                                                                             "nie udało się wygenerować certyfikatu",
	"failed to encode certificate key":                                                                                                           "nie udało się zakodować klucza certyfikatu",
	"failed to save certificate and key":                                                                                                         "nie udało się zapisać certyfikatu i klucza",
	"failed to generate PKCS#12":                                                                                                                 "nie udało się wygenerować PKCS#12",
	"failed to save PKCS#12":                                                                                                                     "nie udało się zapisać PKCS#12",
	"failed to generate serial number":                                                                                                           "nie udało się wygenerować numeru seryjnego",
	"failed to read the CSR":                                                                                                                     "nie udało się odczytać CSR",
	"failed to parse the CSR":                                                                                                                    "nie udało się przetworzyć CSR",
	"invalid CSR signature":                                                                                                                      "nieprawidłowy podpis CSR",
	"failed to parse generated certificate":                                                                                                      "nie udało się przetworzyć wygenerowanego certyfikatu",
	"failed to save certificate":                                                                                                                 "nie udało się zapisać certyfikatu",
	"failed to read the CA certificate":                                                                                                          "nie udało się odczytać certyfikatu CA",
	"failed to parse the CA certificate":                                                                                                         "nie udało się przetworzyć certyfikatu CA",
	"failed to read the CA key":                                                                                                                  "nie udało się odczytać klucza CA",
	"failed to parse the CA key":                                                                                                                 "nie udało się przetworzyć klucza CA",
	"failed to generate the CA key":                                                                                                              "nie udało się wygenerować klucza CA",
	"failed to encode public key":                                                                                                                "nie udało się zakodować klucza publicznego",
	"failed to decode public key":                                                                                                                "nie udało się odkodować klucza publicznego",
	"failed to generate CA certificate":                                                                                                          "nie udało się wygenerować certyfikatu CA",
	"failed to encode CA key":                                                                                                                    "nie udało się zakodować klucza CA",
	"failed to save CA key":                                                                                                                      "nie udało się zapisać klucza CA",
	"failed to save CA certificate":                                                                                                              "nie udało się zapisać certyfikatu CA",
	"failed to read root certificate":                                                                                                            "nie udało się odczytać certyfikatu głównego",
	"failed to create temp file":                                                                                                                 "nie udało się utworzyć pliku tymczasowego",
	"failed to read trust settings":                                                                                                              "nie udało się odczytać ustawień zaufania",
	"failed to parse trust settings":                                                                                                             "nie udało się przetworzyć ustawień zaufania",
	"failed to serialize trust settings":                                                                                                         "nie udało się zakodować ustawień zaufania",
	"failed to write trust settings":                                                                                                             "nie udało się zapisać ustawień zaufania",
	"invalid PEM data":                                                                                                                           "nieprawidłowe dane PEM",
	"decode pem":                                                                                                                                 "dekodowanie PEM",
	"open root store":                                                                                                                            "otwieranie głównego magazynu zaufania",
	"add cert":                                                                                                                                   "dodawanie certyfikatu",
	"delete cert":                                                                                                                                "usuwanie certyfikatu",
	"no certs found":                                                                                                                             "nie znaleziono certyfikatów",
	"failed to open windows root store: %v":                                                                                                      "nie udało się otworzyć głównego magazynu Windows: %v",
	"failed to close windows root store: %v":                                                                                                     "nie udało się zamknąć głównego magazynu Windows: %v",
	"failed adding cert: %v":                                                                                                                     "nie udało się dodać certyfikatu: %v",
	"failed enumerating certs: %v":                                                                                                               "nie udało się wyliczyć certyfikatów: %v",
	"failed duplicating context: %v":                                                                                                             "nie udało się powielić kontekstu: %v",
	"failed deleting certificate: %v":                                                                                                            "nie udało się usunąć certyfikatu: %v",
	"cannot check incomplete PEM file %q for cleanup: %w":                                                                                        "nie można sprawdzić niekompletnego pliku PEM %q przed usunięciem: %w",
	"PEM target %q changed; refusing to remove it during cleanup":                                                                                "plik docelowy PEM %q został zmieniony; odmowa usunięcia podczas wycofywania zapisu",
	"failed to remove incomplete PEM file %q: %w":                                                                                                "nie udało się usunąć niekompletnego pliku PEM %q: %w",
	"refusing to overwrite existing PEM target %q: %w":                                                                                           "odmowa nadpisania istniejącego pliku docelowego PEM %q: %w",
	"failed to create PEM file %q: %w":                                                                                                           "nie udało się utworzyć pliku PEM %q: %w",
	"cannot identify newly created PEM file %q; file left in place: %w":                                                                          "nie można ustalić tożsamości nowego pliku PEM %q; plik pozostawiono na miejscu: %w",
	"failed to save PEM file %q: %w":                                                                                                             "nie udało się zapisać pliku PEM %q: %w",
}

func displayDate(date time.Time) string {
	if uiLanguage == "pl" {
		months := [...]string{"stycznia", "lutego", "marca", "kwietnia", "maja", "czerwca", "lipca", "sierpnia", "września", "października", "listopada", "grudnia"}
		return fmt.Sprintf("%d %s %d", date.Day(), months[date.Month()-1], date.Year())
	}
	return date.Format("2 January 2006")
}

// Keep flag's parsing rules and exit codes, while localizing its own diagnostics.
type languageFlagWriter struct {
	destination io.Writer
	language    *string
}

func (w languageFlagWriter) Write(data []byte) (int, error) {
	text := string(data)
	if *w.language == "pl" {
		for _, pair := range [][2]string{
			{"bad flag syntax: ", "nieprawidłowa składnia parametru: "},
			{"flag provided but not defined: ", "nieznany parametr: "},
			{"flag needs an argument: ", "parametr wymaga wartości: "},
			{"invalid value ", "nieprawidłowa wartość "},
			{"invalid boolean value ", "nieprawidłowa wartość logiczna "},
		} {
			if strings.HasPrefix(text, pair[0]) {
				text = pair[1] + strings.TrimPrefix(text, pair[0])
				text = strings.ReplaceAll(text, " for flag -", " dla parametru -")
				text = strings.ReplaceAll(text, " for -", " dla parametru -")
				if strings.HasSuffix(text, ": parse error\n") {
					text = strings.TrimSuffix(text, ": parse error\n") + ": błąd składni\n"
				}
				break
			}
		}
	}
	if _, err := io.WriteString(w.destination, text); err != nil {
		return 0, err
	}
	return len(data), nil
}
