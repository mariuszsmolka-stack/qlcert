# QuestLab Cert — audyt techniczny etapu 0

- Data: 24 października 2026 r.
- Repozytorium: `C:\questlab-cert`
- Gałąź: `audit/stage0`
- Badany commit: `fcf010598d18f4cad4ac22fd1244d55cfe8c08ff`
- Podstawa zakresu: zatwierdzony [ROADMAP.md](../ROADMAP.md), szczególnie etap 0 i wymagania bezpieczeństwa.

## 1. Wniosek i granice weryfikacji

Kod stanowi niewielką bazę mkcert z działającą kompilacją, ale nie spełnia jeszcze wymagań bezpiecznego eksportu i weryfikacji wydania QuestLab Cert. Najważniejsze problemy to stałe hasło i stare algorytmy PFX, brak zabezpieczeń przed nadpisaniem plików, zbyt szerokie dopasowanie certyfikatów przy usuwaniu zaufania Windows oraz brak testów. `govulncheck` wskazał trzy znane podatności z osiągalnymi symbolami, w tym jedną zależną od użytego kompilatora Go.

Nie potwierdzono problemu o priorytecie **krytycznym**. Nie jest to deklaracja braku takich problemów: nie wykonywano testów penetracyjnych, operacji na magazynach zaufania ani prób ujawnienia kluczy.

Audyt obejmował wszystkie siedem plików Go, `go.mod`, `go.sum`, oba workflow GitHub Actions, `README.md`, `LICENSE`, `AUTHORS` i plan rozwoju. Początkowy stan Git był czysty. Nie uruchamiano programu QuestLab Cert/mkcert ani jego funkcji generowania, instalacji lub usuwania. Nie generowano kluczy CA ani certyfikatów. Narzędzia pobrały metadane modułów, kod skanera i bazę podatności do cache; nie zmieniono zależności projektu. Pliki kompilacji i cache budowania umieszczono w katalogu tymczasowym poza repozytorium.

**Interpretacja ustaleń:**

- **B — potwierdzone błędy lub ograniczenia:** wynikają bezpośrednio z kodu albo uruchomionej kontroli; dla zachowania aplikacji jest to potwierdzenie statyczne, jeśli nie napisano inaczej.
- **R — potencjalne zagrożenia:** warunki ryzyka są widoczne w kodzie, ale skutku w konkretnym środowisku nie odtworzono.
- **M — zalecenia modernizacyjne:** utrzymanie, testowalność i przygotowanie wydania; nie oznaczają automatycznie podatności.

Priorytet **wysoki** oznacza ochronę kluczy, integralności plików lub zaufania albo pilną aktualizację bezpieczeństwa; **średni** — poprawność i niezawodność; **niski** — utrzymanie i porządkowanie. Priorytety są oceną dla lokalnego narzędzia laboratoryjnego, nie oceną CVSS.

## 2. Struktura projektu i przepływ danych

Projekt ma **jeden moduł Go** (`filippo.io/mkcert`, `go.mod:1`) i **jeden pakiet `main`**, a nie osobne moduły aplikacyjne. Podział funkcjonalny istnieje na poziomie plików:

| Plik i lokalizacja | Odpowiedzialność i zależności |
| --- | --- |
| `main.go:87–149,155–238` | Flagi CLI, obiekt `mkcert`, inicjalizacja CA, walidacja nazw, wybór generowania lub operacji zaufania. Wywołuje `cert.go` i adaptery magazynów. |
| `main.go:240–355,358–392` | Lokalizacja CA, koordynacja system/NSS/Java, kontrola zaufania, błędy i procesy z `sudo`. |
| `cert.go:50–146,166–206` | Klucze, X.509, SAN, nazwy plików, PEM i PFX. Korzysta ze standardowej biblioteki kryptograficznej oraz SSLMate PKCS#12. |
| `cert.go:209–367` | Podpisywanie CSR, odczyt i tworzenie CA, identyfikator instalacji oparty na numerze seryjnym. |
| `truststore_windows.go:34–142` | Systemowy adapter Windows przez CryptoAPI, `syscall` i `unsafe`. |
| `truststore_linux.go:27–98` | Wybór katalogu dystrybucji, kopiowanie publicznego CA i odświeżenie magazynu poleceniem systemowym. |
| `truststore_darwin.go:24–110` | Adapter macOS przez `security` i `howett.net/plist`; wybierany przy kompilacji dla Darwin. |
| `truststore_nss.go:39–150` | Wykrywanie profili Firefox/Chromium, program `certutil`, instalacja i usuwanie w NSS. |
| `truststore_java.go:31–121` | `JAVA_HOME`, magazyn `cacerts`, `keytool` i jego hasło. Kod kompilowany także dla Windows/Linux. |

Pliki z sufiksami `_windows.go`, `_linux.go`, `_darwin.go` wybiera Go dla odpowiedniego systemu. NSS i Java są współdzielone; ich dostępność zależy od systemu, ścieżek i zewnętrznych narzędzi. Nie ma serwera CA, Dockera ani bazy danych. `Run` ładuje lub tworzy CA przed sprawdzeniem docelowej operacji i nazw. Podpisywanie odbywa się lokalnie; instalacja zaufania jest wywoływana przez jawne `-install`, a usuwanie przez `-uninstall`.

## 3. Wykonane kontrole

Go nie było dostępne w początkowym `PATH`; odnaleziono `C:\Program Files\Go\bin\go.exe`. `go version` zwróciło **`go version go1.27.0 windows/amd64`**. W kontrolach ustawiono `GOTOOLCHAIN=local`, a `GOCACHE` na `%TEMP%\questlab-stage0-go-cache`. Dla operacji na projekcie używano `-mod=readonly`, gdzie polecenie je obsługuje. Kontrole wymagające zapisu do istniejącego cache modułów powtórzono z przyznanym dostępem po komunikatach `Access is denied` z sandboxa.

| Kontrola | Faktyczny wynik i ograniczenie |
| --- | --- |
| `git branch --show-current`, początkowe `git status --short` | `audit/stage0`; brak zmian. |
| `go list -mod=readonly -m -json all` | Kod 0 po ponowieniu z dostępem do cache; odczytano graf modułów. |
| `go list -mod=readonly -m -u -json all` | Kod 0; odczytano dostępne aktualizacje, bez zmiany `go.mod`/`go.sum`. |
| `go list -mod=readonly -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./...` | Kod 0 po ponowieniu; ustalono rzeczywiście kompilowane pakiety Windows. |
| `go mod verify` | Kod 0, `all modules verified`. Potwierdza integralność zawartości pobranych modułów, nie ich bezpieczeństwo. |
| `go vet -mod=readonly ./...` — Windows/amd64 | Kod 0, bez ostrzeżeń. |
| `go test -mod=readonly ./...` — Windows/amd64 | Kod 0, `? filippo.io/mkcert [no test files]`; nie stanowi pokrycia zachowania testami. |
| `go build -mod=readonly -o <plik w TEMP> ./...` — Windows/amd64 | Kod 0, po ponowieniu bez ostrzeżenia o cache. Plik nie był uruchamiany. |
| `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod=readonly -o <plik w TEMP> ./...` | Kod 0 po ponowieniu z dostępem do cache. Wyłącznie kompilacja krzyżowa, bez testu działania na Linux. |
| `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go vet -mod=readonly ./...` | Kod 0, bez ostrzeżeń. |
| `go run golang.org/x/vuln/cmd/govulncheck@latest -show verbose ./...` | Pobrano skaner `golang.org/x/vuln v1.9.0`. Skan Windows zakończył się wykryciem podatności: narzędzie zgłosiło `exit status 3`, a opakowanie `go run` zwróciło kod 1. Wyniki opisano poniżej. |
| `gofmt -l` dla siedmiu plików Go | Wymienił wszystkie siedem plików. Podgląd `gofmt -d main.go` pokazał różnicę całego pliku; nie stosowano formatowania ani nie traktowano tego jako błędu działania. |
| Odczyt nagłówków siedmiu plików Go i plików licencyjnych | Nagłówki mkcert i odwołania do `LICENSE` zachowane. |

Nie uruchamiano `staticcheck` lokalnie (brak zainstalowanego narzędzia), `go test -race`, testów aplikacji, buildów pozostałych architektur ani wykonania workflow na GitHub. Nie sprawdzono rzeczywistych ACL plików generowanych przez aplikację, importu PFX w IIS/OpenSSL, uprawnień `sudo`, profili przeglądarek, działania Java ani instalacji/usuwania na Windows/Linux. Nie przenosi się wyników historycznych zapisanych w roadmapie na te niewykonane kontrole.

## 4. Wersje i kompatybilność zależności

Wersje dostępne w dniu audytu ustalono przez `go list -m -u -json all`. Są kandydatami do przeglądu, nie zatwierdzonym zestawem aktualizacji. Nie sprawdzono kompilacji projektu po ich zastosowaniu.

| Zależność i lokalizacja | Wersja projektu | Dostępna wersja | Rola / kompatybilność |
| --- | --- | --- | --- |
| Go, `go.mod:3` | deklaracja `1.18`; faktycznie użyto `1.27.0` | w linii 1.27 opublikowano `1.27.3` | Deklaracja w module nie oznacza używania kompilatora 1.18. Go 1.18 nie jest już wspierane; użyte 1.27.0 nie ma najnowszych poprawek. |
| `golang.org/x/net`, `go.mod:6` | `v0.0.0-20220421235706-1d1ef9303861` | `v0.61.0` | Bezpośrednio `idna` w `main.go:27,227`. |
| `howett.net/plist`, `go.mod:7` | `v1.0.0` | `v1.0.1` | Wyłącznie adapter macOS (`truststore_darwin.go:15`); nieobjęty skanem pakietów Windows. |
| `software.sslmate.com/src/go-pkcs12`, `go.mod:8` | `v0.2.0` | `v0.7.4` | Bezpośredni eksport `cert.go:126`. Aktualizacja samego modułu nie zastępuje wyboru bezpiecznego enkodera. |
| `golang.org/x/crypto`, `go.mod:12` | `v0.0.0-20220331220935-ae2d96664a29` | `v0.58.0` | Pośrednio przez PKCS#12; rzeczywiście kompilowany `pbkdf2`. Nie używa się SSH/OpenPGP. |
| `golang.org/x/text`, `go.mod:13` | `v0.3.7` | `v0.43.0` | Pośrednio przez IDNA: normalizacja Unicode, bidi i transformacje. |

W grafie `go list -m all` występują też `x/sys` (`v0.0.0-20211216021012-1d35b9e2eb4e`, dostępne `v0.49.0`), `x/term` (`v0.0.0-20210927222741-03fcf44c2211`, dostępne `v0.47.0`), `x/tools` (`v0.0.0-20180917221912-90fa682c2a6e`, dostępne `v0.52.0`), `go-flags` (`v1.4.0`, dostępne `v1.6.1`), `check.v1` (`v0.0.0-20161208181325-20d25e280405`, dostępna rewizja `v1.0.0-20201130134442-10cb98267c6c`) i `yaml.v1` (`v1.0.0-20140924161607-9f9df34309c0`, bez wskazanej aktualizacji). Obecność w grafie lub wpisu `/go.mod` w `go.sum:1–20` nie oznacza, że kod modułu jest dołączony do programu. Ostatnie trzy pozycje wynikają z zależności narzędzi/testów `plist`; żadna z tych sześciu pozycji nie wystąpiła na liście zewnętrznych pakietów kompilacji Windows.

Aktualne `x/net v0.61.0` deklaruje **Go 1.26.0** i nowsze `x/*`; `go-pkcs12 v0.7.4` deklaruje **Go 1.19**. Utrzymanie deklaracji Go 1.18 przy bezwarunkowym podniesieniu wszystkich wersji byłoby niespójne. Obecny kod kompiluje się użytym Go 1.27.0; minimalnego Go 1.18 oraz nowych zestawów zależności nie testowano. Źródła: [polityka wsparcia i wydania Go](https://go.dev/doc/devel/release), [go.mod x/net v0.61.0](https://raw.githubusercontent.com/golang/net/v0.61.0/go.mod), [go.mod PKCS#12 v0.7.4](https://raw.githubusercontent.com/SSLMate/go-pkcs12/v0.7.4/go.mod).

## 5. Potwierdzone błędy i ograniczenia

### B01 — PFX ma wspólne, ujawniane hasło i zbyt szeroki tryb zapisu

- **Lokalizacja:** `cert.go:125–129,140–142`.
- **Problem:** eksport zawsze używa `changeit`, wypisuje to hasło w logu i zapisuje pakiet zawierający klucz prywatny z trybem `0644`. Na Unix tryb pozwala na odczyt grupie i innym użytkownikom, o ile nie ograniczą tego umask i uprawnienia katalogów. Na Windows nie stanowi gwarancji ochrony ACL.
- **Znaczenie:** posiadacz pliku zna hasło do klucza; aktualny eksport nie spełnia etapu 3 i nie może być udostępniony w kreatorze etapu 2.
- **Priorytet:** wysoki.
- **Rozwiązanie:** do wdrożenia bezpiecznej ścieżki pozostawić kreator przy PEM; następnie wymagać niepustego hasła użytkownika wprowadzanego bez echa, bez jawnego argumentu CLI, historii poleceń i logów. Stosować prywatne uprawnienia i weryfikację ACL. Ten audyt nie wprowadza tej zmiany.

### B02 — eksport PKCS#12 używa parametrów legacy

- **Lokalizacja:** `go.mod:8`, `cert.go:126`; kod przypiętego modułu `software.sslmate.com/src/go-pkcs12@v0.2.0`: `pkcs12.go:454–460,520–525,729–736`, `safebags.go:56–66`, `crypto.go:44–69`, `mac.go:59–65`.
- **Problem:** kod zależności potwierdza RC2-40-CBC dla certyfikatów, 3DES-CBC dla klucza, KDF PKCS#12 oparty na SHA-1 z 2048 iteracjami szyfrowania i 8-bajtowymi solami oraz HMAC-SHA-1 z jedną iteracją KDF MAC. Nie są to parametry nowoczesnego eksportu. RC2 dotyczy certyfikatów, a nie bezpośredniego szyfrowania klucza.
- **Znaczenie:** zmiana samego hasła nie kończy prac nad bezpieczeństwem PFX; możliwe problemy importu w narzędziach wyłączających stare algorytmy.
- **Priorytet:** wysoki.
- **Rozwiązanie:** ocenić nowszą wersję SSLMate i jawny profil PBES2/PBKDF2/AES oraz MAC, dobrać parametry i przetestować import w docelowych Windows/IIS oraz OpenSSL. Nie wybierać profilu wyłącznie według nazwy `Modern`; zgodność importerów wymaga testów. Nowsze API nadal mapuje stare `Encode` na profil legacy, więc sama aktualizacja modułu nie wystarczy. [Dokumentacja biblioteki](https://pkg.go.dev/software.sslmate.com/src/go-pkcs12#Encode).

### B03 — zapis nadpisuje pliki i nie chroni istniejących uprawnień

- **Lokalizacja:** `cert.go:115–129,270–272,355–361`.
- **Problem:** `ioutil.WriteFile` zapisuje bez potwierdzenia i bez wyłącznego tworzenia pliku. Tryb podany przy zapisie nie naprawia uprawnień istniejącego pliku. Zapis certyfikatu i klucza następuje osobno; awaria drugiego zapisu może pozostawić niespójną parę.
- **Znaczenie:** można utracić działający certyfikat lub klucz; zastany plik klucza może pozostać czytelny dla innych. Jest to niezgodne z planowanym potwierdzaniem nadpisania.
- **Priorytet:** wysoki.
- **Rozwiązanie:** oddzielić walidację ścieżek od zapisu, wymagać zgody na nadpisanie, sprawdzać tożsamość plików i uprawnienia, przygotować bezpieczny zapis tymczasowy oraz kontrolowane zatwierdzenie pary. Uwzględnić różnice Windows/Unix. [Semantyka WriteFile](https://pkg.go.dev/os#WriteFile).

### B04 — operacja usuwania i błędne dane mogą tworzyć CA

- **Lokalizacja:** `main.go:171–186,211–237`, `cert.go:282–285,310–361`.
- **Problem:** `Run` tworzy katalog i wywołuje `loadCA` przed rozpatrzeniem `-uninstall`, braku nazw oraz ich poprawności; brak `rootCA.pem` powoduje `newCA`. Nie jest to instalacja zaufania, ale jest zapisem nowego materiału CA.
- **Znaczenie:** administrator próbujący usunąć stare zaufanie przy utraconym certyfikacie dostanie nowe CA zamiast identyfikacji starego. Polecenie z niepoprawną nazwą także może utworzyć CA.
- **Priorytet:** średni.
- **Rozwiązanie:** rozdzielić odczyt i tworzenie CA; zweryfikować operację i dane przed skutkami ubocznymi; usuwanie ma działać bez generowania kluczy i jasno zgłaszać brak identyfikatora CA.

### B05 — Windows usuwa wszystkie certyfikaty o tym samym serialu

- **Lokalizacja:** `truststore_windows.go:53–64,106–139`, szczególnie `130`.
- **Problem:** dopasowanie używa wyłącznie numeru seryjnego; nie porównuje DER, odcisku ani wystawcy. Inne CA mogą mieć ten sam serial. Dodatkowo brak pasującego certyfikatu jest błędem, więc ponowne usunięcie nie jest idempotentne.
- **Znaczenie:** możliwe usunięcie niezwiązanego certyfikatu zaufania. Potwierdzono zbyt szeroki warunek w kodzie; żadnego certyfikatu nie usuwano i nie potwierdzono kolizji na tym komputerze.
- **Priorytet:** wysoki.
- **Rozwiązanie:** identyfikować dokładny certyfikat np. przez DER/odcisk SHA-256, przetestować dwa różne certyfikaty o tym samym serialu w atrapach i traktować brak docelowego wpisu jako jawny stan bez zmian.

### B06 — CSR bez obsługiwanych SAN może wywołać panic

- **Lokalizacja:** `cert.go:228–239,259–268,176–177`.
- **Problem:** `hosts` jest budowane z odczytanych SAN; `fileNames` bezwarunkowo odczytuje `hosts[0]`. CSR z rozszerzeniem SAN bez DNS/IP/email/URI może zastąpić domyślny SAN i pozostawić pustą listę. Własne `-cert-file` nie pomaga, bo domyślna nazwa jest obliczana wcześniej.
- **Znaczenie:** wejście CSR może zakończyć program panic zamiast czytelnym błędem. To statyczna analiza ścieżki; nie generowano CSR ani nie wykonywano podpisywania.
- **Priorytet:** średni.
- **Rozwiązanie:** sprawdzać obecność obsługiwanych tożsamości przed podpisywaniem i tworzeniem nazw, uniezależnić jawną nazwę pliku od SAN, dodać regresję dla pustego/nietypowego SAN.

### B07 — błąd instalacji może zostać zamaskowany

- **Lokalizacja:** `main.go:267–305,336–342`, szczególnie `272–275`; `truststore_linux.go:55–59`.
- **Problem:** `ignoreCheckFailure` jest ustawiane także wtedy, gdy `installPlatform` zwróci `false`. Dla nieobsługiwanej dystrybucji Linux taka ścieżka istnieje jawnie. Braki narzędzi NSS/Java są ostrzeżeniami i nie składają się na wiarygodny wynik całej operacji.
- **Znaczenie:** powodzenie polecenia i późniejszy stan kontroli w tym procesie nie dowodzą ustanowienia zaufania we wszystkich żądanych magazynach.
- **Priorytet:** średni.
- **Rozwiązanie:** zwracać ustrukturyzowany wynik per magazyn, odróżnić pominięcie, brak obsługi i błąd; po operacji sprawdzać docelowy stan, a nie bezwarunkowo pomijać kontrolę.

### B08 — znane podatności w osiągalnej ścieżce IDNA/Unicode

- **Lokalizacja:** `go.mod:6,13`, `main.go:227` (`idna.ToASCII`).
- **Problem:** `govulncheck v1.9.0` wskazał [GO-2026-5970](https://pkg.go.dev/vuln/GO-2026-5970) w normalizacji `x/text` (naprawa `v0.39.0`) i [GO-2026-5026](https://pkg.go.dev/vuln/GO-2026-5026) w IDNA `x/net` (naprawa `v0.55.0`). Pierwszy wpis dotyczy zapętlenia przy niepoprawnym UTF-8, drugi akceptowania niepoprawnego Punycode. Skan wskazał ścieżki z `Run` do podatnych symboli.
- **Znaczenie:** walidacja nazw certyfikatu opiera się na podatnych wersjach. Osiągalność statyczna nie potwierdza wykonalnego ataku w tym CLI; nie odtwarzano zawieszenia, a program nie ma opisanego w advisory systemu autoryzacji domen.
- **Priorytet:** wysoki.
- **Rozwiązanie:** dobrać aktualny, zgodny zestaw `x/net` i `x/text` wraz z Go; dodać przypadki niepoprawnego UTF-8/Punycode i ponowić skan. Podane wersje naprawiają konkretne wpisy, nie wszystkie późniejsze podatności modułów.

### B09 — użyty kompilator nie zawiera poprawki weryfikacji X.509

- **Lokalizacja:** `main.go:341`; środowisko audytu Go `1.27.0`, workflow `.github/workflows/test.yml:12–15`, `release.yml:10–13`.
- **Problem:** skaner wskazał [GO-2026-6809](https://pkg.go.dev/vuln/GO-2026-6809) przez `checkPlatform → x509.Certificate.Verify`; poprawka w linii 1.27 jest w `1.27.3`. Dotyczy dodatkowego odrzucania określonych skompromitowanych certyfikatów pośrednich. Wersja kompilatora nie jest przypięta przez projekt.
- **Znaczenie:** artefakt zbudowany lokalnym Go 1.27.0 dziedziczy ten brak. Nie potwierdzono, że lokalny CA tworzy taki łańcuch, ani że istniejące opublikowane pliki projektu powstały tą wersją Go.
- **Priorytet:** średni.
- **Rozwiązanie:** budować aktualnym poprawionym Go, zapisywać wersję narzędzi w wynikach CI i skanować artefakty. Sama zmiana `go 1.18` nie aktualizuje automatycznie bibliotek standardowych w istniejących binariach.

## 6. Potencjalne zagrożenia wymagające sprawdzenia

### R01 — uprawnienia kluczy na Windows i zastane katalogi

- **Lokalizacja:** `main.go:176,240–264`, `cert.go:113–121,355–356`.
- **Problem:** CA i klucze PEM są nieszyfrowanym PKCS#8. `0400`/`0600` ograniczają nowe pliki na Unix, lecz na Windows nie ustanawiają prywatnej DACL. `CAROOT` może wskazać dowolny, już istniejący katalog; jego właściciel i uprawnienia nie są sprawdzane, nowy katalog ma tryb `0755`.
- **Znaczenie:** wyciek klucza CA pozwala wystawiać certyfikaty zaufane klientom laboratorium. Nie ustalono faktycznych uprawnień plików użytkownika i nie stwierdza się, że domyślny profil Windows jest publicznie czytelny.
- **Priorytet:** wysoki.
- **Rozwiązanie:** ustalić i testować ochronę katalogu oraz plików osobno dla Windows i Linux, odrzucać niebezpieczne lokalizacje lub wymagać świadomej decyzji, opisać ochronę kopii zapasowych. [Ograniczenia trybów plików Windows w Go](https://pkg.go.dev/os#Chmod).

### R02 — nieatomowe tworzenie CA, dowiązania i równoległe uruchomienia

- **Lokalizacja:** `cert.go:282–285,310–361`, `main.go:370–373`; zapis eksportów `cert.go:115–129`.
- **Problem:** brak blokady CA i transakcji klucz/certyfikat; `pathExists` uznaje każdy błąd `Stat` za brak pliku. Zapisy nie odrzucają dowiązań ani aliasów ścieżek. Brak certyfikatu przy zachowanym kluczu uruchamia próbę utworzenia nowej pary; zależnie od praw może ona nadpisać klucz albo zakończyć się błędem.
- **Znaczenie:** przerwanie zapisu lub wyścig może pozostawić uszkodzony stan CA; w katalogu kontrolowanym przez inną osobę istnieje ryzyko skierowania zapisu poza zamierzone miejsce. Tych sytuacji nie odtwarzano.
- **Priorytet:** wysoki.
- **Rozwiązanie:** rozróżniać brak od błędów dostępu, chronić katalog, weryfikować pliki i dowiązania, blokować inicjalizację, stosować bezpieczne zapisy i jawne odzyskiwanie niekompletnego CA bez automatycznego zastępowania klucza.

### R03 — za mała walidacja istniejącego CA i okresu certyfikatu

- **Lokalizacja:** `cert.go:282–307,62–74,228–245`.
- **Problem:** `loadCA` parsuje PEM/DER, ale nie ma własnego pełnego sprawdzenia pary klucz/certyfikat, `IsCA`, `KeyUsageCertSign`, samopodpisu i ważności. Okres certyfikatu końcowego nie jest ograniczony do `NotAfter` CA, a `NotBefore` jest ustawiane na bieżący czas bez tolerancji rozbieżności zegarów.
- **Znaczenie:** możliwe niejasne błędy biblioteki lub certyfikaty, których łańcuch przestaje działać wcześniej, niż sugeruje komunikat o dacie końcowej. `x509.CreateCertificate` może sam odrzucić część niezgodności; brak własnych kontroli nie dowodzi, że każda błędna para zostanie użyta.
- **Priorytet:** średni.
- **Rozwiązanie:** walidować CA przed operacjami, jawnie obsłużyć wygaśnięcie, ograniczyć ważność certyfikatu do CA i określić politykę tolerancji zegara. Przetestować błędną parę, klucz bez certyfikatu, CA bez klucza oraz certyfikaty wygasłe/przyszłe.

### R04 — CSR przekazuje rozszerzenia bez polityki wystawcy

- **Lokalizacja:** `cert.go:224–245,254`, zwłaszcza `ExtraExtensions: csr.Extensions` w `232`.
- **Problem:** podpis CSR jest poprawnie sprawdzany, lecz wszystkie żądane rozszerzenia są kopiowane, bez listy dozwolonych OID, kontroli żądanych zastosowań i własnej walidacji klucza publicznego.
- **Znaczenie:** żądanie może narzucić rozszerzenia inne niż zamierzony profil certyfikatu HTTPS. Nie potwierdzono obejścia walidacji klientów ani skutecznego utworzenia podrzędnego CA; nowy root ma `MaxPathLenZero`, co również ogranicza łańcuch.
- **Priorytet:** wysoki.
- **Rozwiązanie:** zdefiniować politykę CSR: dozwolone SAN, algorytmy i rozmiary kluczy, KU/EKU, odrzucenie uprawnień CA i nieobsługiwanych krytycznych rozszerzeń. Przetestować ją na kontrolowanych danych bez używania rzeczywistego CA.

### R05 — nazwa pliku zależy od nieograniczonej tożsamości wejściowej

- **Lokalizacja:** `main.go:216–234`, `cert.go:176–199,259–268`.
- **Problem:** domyślna nazwa zastępuje tylko `:` i `*`. Akceptowane URI mogą zawierać separatory ścieżek; SAN z CSR nie przechodzą tej samej walidacji co CLI. Porównanie `certFile == keyFile` jest tekstowe i nie rozpoznaje dwóch nazw tego samego pliku.
- **Znaczenie:** błędne ścieżki, niezamierzony katalog docelowy albo zapis klucza na certyfikat przy aliasach plików. Nie wykonano próby traversal ani zapisu na cudzy plik.
- **Priorytet:** średni.
- **Rozwiązanie:** tworzyć bezpieczną nazwę bazową niezależnie od treści SAN, sprawdzać wynikową ścieżkę i tożsamość plików; oddzielić nazwę wyjściową od walidacji tożsamości certyfikatu.

### R06 — zakres zaufania Windows może być niejasny dla użytkownika

- **Lokalizacja:** `truststore_windows.go:70–79`, `main.go:269–273`, `README.md:119–135`.
- **Problem:** `CertOpenSystemStoreW("ROOT")` otwiera magazyn bieżącego użytkownika; nie wybiera jawnie `LocalMachine`. Komunikat mówi ogólnie o magazynie systemowym. Firefox/NSS Windows nie jest obsługiwany przez ten adapter (`truststore_windows.go:20–22`, `truststore_nss.go:48–70`).
- **Znaczenie:** zaufanie konta administratora nie musi oznaczać zaufania innego klienta, konta usługi czy przeglądarki. Nie sprawdzano zachowania konkretnych usług ani polityk Windows.
- **Priorytet:** średni.
- **Rozwiązanie:** opisywać dokładny zakres instalacji i konto; w dokumentacji laboratorium osobno weryfikować klienta, serwer i magazyn używany przez daną aplikację. Nie rozszerzać automatycznie instalacji na cały komputer. [Dokumentacja Microsoft](https://learn.microsoft.com/en-us/windows/win32/api/wincrypt/nf-wincrypt-certopensystemstorew).

### R07 — Linux i NSS: częściowe operacje oraz usuwanie starego zaufania

- **Lokalizacja:** `truststore_linux.go:36–47,65–74,82–96`; `truststore_nss.go:106–115`; `main.go:307–333`.
- **Problem:** zapis/usunięcie pliku i odświeżenie magazynu Linux są osobnymi krokami bez rollbacku. Legacy `mkcert-rootCA` jest usuwany po nazwie, bez porównania certyfikatu. NSS przed usunięciem uruchamia `certutil -V`; każdy błąd pomija `-D`, również gdy wpis istnieje, ale nie przechodzi walidacji. Komunikat zbiorczy może nie odzwierciedlać wszystkich pominięć.
- **Znaczenie:** możliwe pozostawienie starego wpisu lub częściowej zmiany zaufania. Nie testowano dystrybucji, baz NSS ani wygaśniętego CA. `-V` jest kontrolą ważności, nie samą kontrolą istnienia — [dokumentacja NSS](https://nss-crypto.org/reference/security/nss/legacy/reference/nss_tools__colon__certutil/index.html).
- **Priorytet:** średni.
- **Rozwiązanie:** usuwać dokładnie zidentyfikowany certyfikat niezależnie od jego aktualnej ważności, rozróżniać błąd dostępu i brak wpisu; zwracać wyniki każdego kroku i opracować odzyskiwanie po nieudanym odświeżeniu.

### R08 — procesy zewnętrzne, hasło Java i kończenie przez Fatal

- **Lokalizacja:** `truststore_java.go:28,57–71,81–119`, `truststore_nss.go:79,91–93,120–126`, `main.go:358–392`, `truststore_windows.go:47–49,58–64`.
- **Problem:** hasło Java `changeit` jest podawane w argumentach `keytool`; nie jest to hasło eksportu PFX, ale stanowi wzorzec nieodpowiedni dla nowej obsługi sekretów. Zmienione hasło `cacerts` może przerwać nawet zwykłe generowanie przez `checkJava`. Procesy nie mają limitów czasu; część jest wyszukiwana w `PATH` i może być uruchamiana przez `sudo`. `log.Fatal` kończy proces bez wykonania `defer`; błędy `close()` magazynu Windows są pomijane.
- **Znaczenie:** możliwe zawieszenie, niejasna diagnostyka, ujawnienie sekretu procesu i trudniejsze sprzątanie zasobów. Nie wykazano podmiany narzędzia w `PATH` ani trwałego wycieku uchwytu: system zwalnia zasoby po zakończeniu procesu.
- **Priorytet:** średni.
- **Rozwiązanie:** zwracać błędy do warstwy CLI, zapewnić sprzątanie i diagnostykę zamykania, stosować kontrolowane ścieżki procesów/środowisko i limity czasu; obsłużyć konfigurację Java bez jawnych sekretów w argumentach. Nie kopiować mechanizmu Java do PFX.

## 7. Zalecenia modernizacyjne i braki testów

### M01 — brak testów i granic pozwalających na bezpieczne testowanie

- **Lokalizacja:** cały pakiet: `main.go:155–238,267–392`, `cert.go:50–367`, adaptery magazynów; brak plików `*_test.go`; `.github/workflows/test.yml:22–23`.
- **Ograniczenie:** `go test` tylko kompiluje pakiet. Kryptografia, dysk, CLI i systemowe magazyny są połączone; bez atrap łatwo przypadkowo zmienić zaufanie podczas testu.
- **Znaczenie:** zielone CI nie potwierdza poprawności enumeracji, eksportu, obsługi błędów ani zachowania na docelowym systemie.
- **Priorytet:** wysoki.
- **Rozwiązanie:** na etapie 0 uzgodnić plan i granice testów. Przewidzieć testy DNS/IPv4/IPv6/IDN/wildcard, niepoprawnych wejść i CSR, pustych SAN, par klucz–certyfikat, ważności, KU/EKU, PFX i ACL, istniejących plików, awarii zapisu i wyścigów. Adapter Windows wymaga przypadków końca enumeracji, błędu parsowania, duplikacji/usunięcia i zwolnienia kontekstu oraz kolizji serialu. Linux/NSS/Java: brak narzędzia, błąd polecenia, częściowy sukces i ponowienie. Testy mają używać atrap i katalogów tymczasowych; testy integracyjne wyłącznie w odizolowanym laboratorium, zgodnie z roadmapą.

### M02 — CI nie zapewnia powtarzalnego zestawu kontroli

- **Lokalizacja:** `.github/workflows/test.yml:9–23`, `.github/workflows/release.yml:8–26`.
- **Ograniczenie:** stare główne wersje `setup-go@v2`, `checkout@v2`, `github-script@v3`; ruchome `go-version: 1.x`, systemy `*-latest` i `staticcheck@latest`. Nazwa kroku używa nieistniejącej osi `matrix.go`. Brak jawnego `govulncheck`, kontroli formatowania, `go mod verify` i polityki minimalnych uprawnień tokena. `go test` domyślnie wykonuje część kontroli vet, więc nie należy mówić o całkowitym braku vet, ale nie zastępuje to jawnego `go vet ./...`.
- **Znaczenie:** wynik zależy od dnia uruchomienia; nie ma bramki wykrywającej aktualne podatności ani dowodu działania wymaganej minimalnej wersji Go. Nie sprawdzono historii ani rzeczywistych statusów Actions.
- **Priorytet:** średni.
- **Rozwiązanie:** wybrać wspierany Go i przypięte wersje narzędzi/akcji, zweryfikować wymagania runnerów, ustawić odpowiednie `permissions`, kontrolować vet, podatności i integralność modułów. Zachować Windows jako pierwszy cel i osobną weryfikację Linux; przypisać wyniki do wersji narzędzi. Aktualne wymagania akcji: [setup-go](https://github.com/actions/setup-go), [checkout](https://github.com/actions/checkout).

### M03 — wydanie nie pakuje wymaganych materiałów i nie ma bramki testów

- **Lokalizacja:** `.github/workflows/release.yml:1–3,16–24,25–48`; `LICENSE:7–15`, `AUTHORS:1–11`.
- **Ograniczenie:** workflow po publikacji release buduje i wysyła jedynie pliki zaczynające się od `mkcert-`, bez pakietu `LICENSE`, `AUTHORS` i dokumentacji, sum kontrolnych ani powiązania z pomyślnym jobem testowym. `git describe --tags` zakłada dostępność tagów, której workflow nie konfiguruje jawnie. Domyślny checkout jest płytki; nie odtworzono konkretnego błędu wydania.
- **Znaczenie:** samodzielny pobrany plik binarny nie ma dołączonych materiałów wymaganych przez roadmapę; warunek licencji dla dystrybucji binarnej trzeba obsłużyć w materiałach wydania. Nie zbadano istniejących release i nie stwierdza się na tej podstawie naruszenia prawnego opublikowanej dystrybucji.
- **Priorytet:** średni.
- **Rozwiązanie:** przygotować archiwa z dokumentacją, oryginalnym `LICENSE`, `AUTHORS` i informacjami licencyjnymi zależności; dodać sumy, testy poprzedzające publikację oraz deterministyczną wersję z tagu zdarzenia lub jawne pobranie potrzebnych tagów. Bez publikacji w ramach tego audytu. [Zachowanie checkout](https://github.com/actions/checkout#usage).

### M04 — wersje i dokumentacja odziedziczone po mkcert wymagają uzgodnienia

- **Lokalizacja:** `go.mod:1–14`, `README.md:1–11,83–90,119`, `main.go:30–80,240–264`, `cert.go:67–68,330–336`.
- **Ograniczenie:** README podaje Go 1.13+/1.10+, moduł deklaruje 1.18, brak ustalonej polityki wspieranego kompilatora. Nazwa modułu, pomoc, katalog CA i certyfikaty nadal identyfikują mkcert; README zaczyna się od qlcert. Nie ma jeszcze menu PL/EN, kreatora ani diagnostyki z roadmapy — to planowane funkcje, nie regresje obecnego kodu.
- **Znaczenie:** początkujący użytkownik może budować nieodpowiednim Go albo korzystać z instrukcji instalujących upstream. Zmiana domyślnego katalogu CA w przyszłości wymaga ochrony istniejącego zaufania i migracji.
- **Priorytet:** średni.
- **Rozwiązanie:** w etapie 0 ustalić macierz zgodności i plan migracji zależności/CA; rebranding i PL/EN wykonać w etapie 1. Nie zmieniać zakresu ani kolejności zatwierdzonego planu.

### M05 — małe ograniczenia generowania i utrzymania

- **Lokalizacja:** `cert.go:202–206,325–338`, `main.go:216–234`, `cert.go:18`, `truststore_darwin.go:60–62,76–85`.
- **Ograniczenie:** losowanie serialu z zakresu zaczynającego się od zera teoretycznie może zwrócić zero (prawdopodobieństwo `2^-128`); regex DNS nie stanowi pełnej walidacji etykiet i długości. Pozostało `io/ioutil`. Adapter Darwin nie zamyka jawnie pliku tymczasowego i używa niechronionych rzutowań typów danych plist. `gofmt -l` zgłasza pliki; przed ewentualnym formatowaniem trzeba rozróżnić końce linii od zmian formatowania.
- **Znaczenie:** rzadkie błędy i trudniejsze utrzymanie; problemy Darwin mają mniejszy priorytet niż Windows/Linux. SHA-1 w `SubjectKeyId` jest identyfikatorem klucza, a nie podpisem CA — sama ta obecność nie dowodzi niebezpiecznego algorytmu podpisu.
- **Priorytet:** niski.
- **Rozwiązanie:** gwarantować dodatni serial, określić profil walidacji DNS, modernizować API przy zmianach funkcjonalnych i poprawić obsługę zasobów/typów Darwin przed deklarowaniem jego wsparcia. Nie wykonywano tych zmian w audycie.

## 8. Zweryfikowane właściwości kryptografii i zasobów

Statyczny przegląd potwierdza następujące właściwości obecnej implementacji; nie zastępują one testów generowania i weryfikacji:

- `cert.go:166–174`: źródłem losowości jest `crypto/rand`; domyślnie RSA 3072 dla CA i RSA 2048 dla certyfikatu, opcjonalnie ECDSA P-256 również dla nowego CA. Nie znaleziono własnego generatora losowego ani ręcznej implementacji podpisów.
- `cert.go:327–350`: root jest samopodpisany, ma `IsCA`, `BasicConstraintsValid`, `KeyUsageCertSign`, `MaxPathLenZero` i 10-letnią ważność. `MaxPathLenZero` odpowiada CA wystawiającemu bezpośrednio certyfikaty końcowe.
- `cert.go:62–104`: certyfikaty otrzymują SAN dla DNS/IP/email/URI i odpowiednie EKU według wejścia; okres wynosi 2 lata i 3 miesiące. Algorytm podpisu wybiera `crypto/x509`, kod nie wymusza SHA-1 do podpisywania. CN w eksporcie PFX jest dodatkiem dla IIS, nie zastępuje SAN.
- `cert.go:214–226`: kontrolowane są typ bloku PEM, parsowanie i podpis CSR. Zastrzeżenia dotyczą polityki rozszerzeń, nie pominięcia weryfikacji podpisu żądania.
- `cert.go:111–126,353–360`: standardowy eksport obejmuje klucz certyfikatu końcowego; PFX zawiera certyfikat CA jako łańcuch, ale **nie klucz prywatny CA**. Klucz CA zapisuje się oddzielnie w `CAROOT`.
- `cert.go:296–298`: brak klucza przy istniejącym publicznym CA dopuszcza tryb bez klucza; wydawanie nowego certyfikatu sprawdza brak klucza (`51–52`, `210–211`).
- `main.go:179–204,267–334`: zwykłe generowanie samo nie instaluje zaufania; `-install` jest jawnym poleceniem CLI. Dodatkowa zgoda kreatora należy do przyszłych etapów.
- `truststore_windows.go:106–142`: poprawka używa `syscall.CertEnumCertificatesInStore`, rozpoznaje `CRYPT_E_NOT_FOUND`, duplikuje kontekst przed usunięciem i zwalnia bieżący kontekst przy wcześniejszym powrocie. Nie znaleziono powrotu do poprzedniego problemu rzutowania wskaźnika; świeży `go vet` jest czysty. Nie wykonano enumeracji na rzeczywistym magazynie.
- Większość błędów kryptografii i I/O jest przekazywana do `fatalIfErr`; wyjątki obejmują m.in. ignorowany błąd parsowania własnego certyfikatu w `cert.go:125`, celowo pomijane błędy parsowania podczas enumeracji (`truststore_windows.go:128–130`), błędy `Stat` oraz zamykania. Brak ostrzeżeń vet nie stanowi pełnego sprawdzenia tych ścieżek.

## 9. Zakres wyniku skanowania podatności

Skan Windows objął pakiet projektu, `x/crypto`, `x/net`, `x/text`, `go-pkcs12` i bibliotekę standardową Go 1.27.0. Wynik: **3 podatności na poziomie osiągalnych symboli**, **2 dodatkowe na poziomie importowanych pakietów** i **55 dodatkowych na poziomie modułów**, dla których skaner nie wskazał osiągalnego wywołania. Nie należy przedstawiać wszystkich 60 wpisów jako 60 potwierdzonych ataków na aplikację.

Dwa wpisy pakietowe to [GO-2026-6608](https://pkg.go.dev/vuln/GO-2026-6608) w `net/textproto` i [GO-2026-6604](https://pkg.go.dev/vuln/GO-2026-6604) w `os` (Windows), z poprawkami w Go 1.27.2. Punkt powiązania z projektem to importy `main.go:15,17`, nie potwierdzona podatna operacja. Dla tej grupy i niższego poziomu modułów: **priorytet średni**, znaczenie — zaległość aktualizacji i ryzyko przy przyszłym rozszerzaniu użycia bibliotek; rozwiązanie — aktualizacja zgodnego zestawu i ponowny skan, bez dodawania SSH/HTTP do zakresu projektu.

Przykładowo [GO-2022-1059](https://pkg.go.dev/vuln/GO-2022-1059) (`x/text/language`) i [GO-2023-1571](https://pkg.go.dev/vuln/GO-2023-1571) (`x/net/http2`) pasują do wersji w `go.mod:6,13`, ale te pakiety nie znalazły się w rzeczywiście kompilowanej ścieżce Windows. Analogicznie liczne alerty SSH w `x/crypto` nie oznaczają, że program obsługuje SSH.

Biblioteka PKCS#12 nie otrzymała alertu skanera dla użytej wersji. Osobno sprawdzony [GO-2026-5052](https://pkg.go.dev/vuln/GO-2026-5052) dotyczy dekodowania w wersjach od `v0.6.0` do wersji przed `v0.7.2`; nie obejmuje `v0.2.0`, a aplikacja używa kodowania. Nie unieważnia to potwierdzonych problemów hasła i algorytmów eksportu.

Nie przeprowadzono osobnego skanu Darwin ani analizy wywołań Linux. Brak alertu dla `plist` nie został zweryfikowany skanem tej platformy. Aktualizacje należy oceniać ponownie w dniu ich wdrożenia.

## 10. Licencja i autorzy

`LICENSE:1–27` zawiera copyright mkcert, trzy warunki redystrybucji i pełne wyłączenie odpowiedzialności. Porównanie treści z [LICENSE upstream mkcert](https://raw.githubusercontent.com/FiloSottile/mkcert/master/LICENSE) nie ujawniło usunięcia tych elementów; nie wykonywano porównania bajtowego z historycznym commitem źródłowym. `AUTHORS:1–11` zawiera oryginalną listę i wyjaśnienie jej charakteru; pełnej zgodności z historią upstream nie zweryfikowano (zdalny odczyt AUTHORS nie powiódł się). Wszystkie siedem plików Go zachowuje nagłówek copyright mkcert i odesłanie do `LICENSE` w liniach 1–3. `README.md:3,11` podaje pochodzenie i BSD-3-Clause.

W źródłach nie znaleziono usunięcia tych informacji. Otwartą pracą jest ich dostarczenie z pakietem binarnym oraz inwentaryzacja licencji zależności (M03). Ten przegląd plików nie stanowi pełnej oceny prawnej dystrybucji ani weryfikacji istniejących release.

## 11. Uporządkowane prace do zakończenia etapu 0

Poniższa lista przekłada kryteria etapu 0 na konkretne rezultaty. Nie przenosi wszystkich funkcji etapów 1–6 do etapu 0 i nie oznacza ich wykonania przez audyt.

1. **Zatwierdzić bazę i rejestr ustaleń.** Zachować commit, wyniki kontroli i ten raport; nadać właścicieli B01–B09, R01–R08 i M01–M05. Rozdzielić błędy do poprawy w bazie od prac wdrożeniowych dalszych etapów. Nie oznaczać etapu 0 jako ukończonego tylko na podstawie kompilacji.
2. **Uzgodnić wspierane Go i plan aktualizacji zależności.** Dobrać poprawioną wersję Go oraz zgodny zestaw `x/*`, sprawdzić minimalne wymagania, API i licencje; ustalić traktowanie zależności używanych tylko przez Darwin/narzędzia. Po przyszłych zmianach wykonać build, vet i skan Windows oraz Linux. Udokumentować wszystkie pozostałe alerty i ich osiągalność (B08–B09, M04).
3. **Zamknąć ocenę PKCS#12 decyzją techniczną.** Udokumentować wybrane API, algorytmy/KDF/MAC, wymagania haseł i macierz docelowych importerów. Wskazać zadania etapu 3; do jego zakończenia kreator etapu 2 ma udostępniać tylko PEM, bez eksportu `changeit` (B01–B02). Implementacja nowego eksportu pozostaje w etapie 3.
4. **Uzgodnić poprawki niezawodności i ochrony zasobów.** Priorytetowo zaplanować dokładne dopasowanie przy usuwaniu Windows, brak tworzenia CA podczas usuwania, obsługę pustych SAN, błędów plików i wyników instalacji (B03–B07). Dla ryzyk R01–R08 określić warunki reprodukcji oraz przypisać poprawki do etapu 0 lub odpowiedniego późniejszego etapu, z uzasadnieniem i kryteriami akceptacji.
5. **Spisać i zatwierdzić plan bezpiecznych testów.** Zaprojektować atrapy magazynów i procesów, dane testowe i katalogi tymczasowe oraz przypadki regresji z M01. Oddzielić testy bez skutków systemowych od późniejszego laboratorium integracyjnego. Wyniku `[no test files]` nie uznawać za test funkcji.
6. **Ustalić powtarzalne kontrole CI.** Określić wersje Go, akcji i analizatorów, kolejność Windows/Linux, wyniki skanowania i minimalne uprawnienia. Przegląd workflow zakończyć listą zmian i sposobem ich sprawdzenia; nie deklarować sukcesu uruchomień, których nie wykonano (M02–M03).
7. **Zamknąć przegląd licencyjny źródeł i plan pakowania.** Zachować `LICENSE`, `AUTHORS` i nagłówki; uzupełnić weryfikację pochodzenia AUTHORS/historycznego punktu bazowego oraz listę licencji zależności. Zaplanować dołączanie materiałów do wydania 1.0 i sprawdzenie kompletności pakietu (M03).
8. **Zapisać decyzję o gotowości bazy.** Potwierdzić wykonane prace etapu 0, udokumentować nierozwiązane kwestie i ich przypisanie do roadmapy. Próby rzeczywistego zaufania, importu PFX, wdrożenia HTTPS oraz końcowy scenariusz początkującego administratora pozostają do odizolowanych testów odpowiednich etapów. W tym audycie nie zostały wykonane.
