# QuestLab Cert — historia rozwoju i rejestr napraw

**Projekt:** QuestLab Cert  
**Projekt źródłowy:** [FiloSottile/mkcert](https://github.com/FiloSottile/mkcert)  
**Repozytorium:** [mariuszsmolka-stack/qlcert](https://github.com/mariuszsmolka-stack/qlcert)  
**Data utworzenia:** 2026-10-24  
**Źródło stanu bazowego:** `docs/TECHNICAL_AUDIT.md`, audyt z 2026-10-24, commit `fcf010598d18f4cad4ac22fd1244d55cfe8c08ff`.

## Cel i zasada aktualizowania

Dokument pokazuje **w kolejności wykonania**, co QuestLab faktycznie poprawił albo dodał względem bazowego `mkcert`. Historia jest również podstawą do rzetelnego publicznego podsumowania wydania 1.0.

1. **Każda zakończona i zatwierdzona zmiana otrzymuje NOWY WPIS NA KOŃCU dokumentu**, w sekcji „Dziennik wykonanych zmian”. Nie wpisujemy tam samych planów ani niepowiązanych commitów roboczych.
2. Każdy wpis podaje: datę, kolejny numer, ID ustalenia audytu (jeśli istnieje), stan przed i po, pliki, numery GO/CVE/GHSA (jeśli dotyczą), commit, PR i wyniki testów. Nie wymyślamy identyfikatorów podatności.
3. Oddzielamy **wykrycie** od **naprawy w kodzie**, **weryfikacji testami** i **wydania użytkownikom**. Wyłączenie ostrzeżenia lub samo podniesienie wersji nie jest bez dowodów potwierdzeniem usunięcia ryzyka.
4. Rejestr ustaleń poniżej może być aktualizowany o stan („otwarte”, „poprawione”, „zweryfikowane”, „odłożone z uzasadnieniem”), ale **chronologii w sekcji końcowej nie przepisujemy i nie usuwamy**. Korektę historycznego wpisu dokumentujemy następnym wpisem.
5. Wpis przygotowujemy na gałęzi danej zmiany po uzyskaniu wyników kontroli i dopisujemy do końca dziennika. Po otwarciu PR uzupełniamy jego odnośnik, jeszcze przed scaleniem. Informacje dostępne dopiero po scaleniu (np. SHA commita scalającego) dopisujemy w późniejszej aktualizacji bez usuwania historii. Końcowe podsumowanie do `README.md` przygotujemy z wpisów potwierdzonych.
6. Zachowujemy zasługi autorów `mkcert` oraz warunki BSD-3-Clause; nie przedstawiamy istniejącego mechanizmu kryptograficznego jako autorskiego dzieła QuestLab.

## Rejestr podatności zidentyfikowanych w audycie

Według skanu `govulncheck` w raporcie: **3 wyniki na poziomie osiągalnego symbolu**, **2 wyniki na poziomie importowanego pakietu** oraz **55 kolejnych na poziomie modułu**. To **nie jest 60 potwierdzonych możliwości zaatakowania aplikacji**. Symbol reachability jest wynikiem analizy statycznej; nie przeprowadzono eksploatacji podatności.

| ID GO | CVE | Dotknięta część | Ocena audytu | Warunek uznania za naprawione | Status |
| --- | --- | --- | --- | --- | --- |
| [GO-2026-5970](https://pkg.go.dev/vuln/GO-2026-5970) | `CVE-2026-56852` | `golang.org/x/text`, Unicode/niepoprawny UTF-8; poprawione od `v0.39.0` | Osiągalny symbol, B08 | Aktualizacja zgodnego zestawu, testy, ponowny `govulncheck` | **Otwarte** |
| [GO-2026-5026](https://pkg.go.dev/vuln/GO-2026-5026) | `CVE-2026-39821` | `golang.org/x/net/idna`, nieprawidłowy Punycode; poprawione od `v0.55.0` | Osiągalny symbol, B08 | Aktualizacja, testy DNS/IDNA, ponowny `govulncheck` | **Otwarte** |
| [GO-2026-6809](https://pkg.go.dev/vuln/GO-2026-6809) | `CVE-2026-111943` | Go `crypto/x509.Certificate.Verify`; poprawione w Go 1.27.3 | Osiągalny symbol, B09 | Kompilacja poprawionym Go, testy i ponowny skan | **Otwarte** |
| [GO-2026-6608](https://pkg.go.dev/vuln/GO-2026-6608) | `CVE-2026-94440` | Go `net/textproto`, limit pamięci podczas parsowania | Pakiet importowany, bez wskazanej osiągalnej operacji | Aktualizacja kompilatora i ponowny skan | **Do weryfikacji po aktualizacji** |
| [GO-2026-6604](https://pkg.go.dev/vuln/GO-2026-6604) | `CVE-2026-56857` | Go `os`, przypadek dowiązania junction w Windows | Pakiet importowany, bez wskazanej osiągalnej operacji | Aktualizacja kompilatora i ponowny skan | **Do weryfikacji po aktualizacji** |

**Wersje podane w tabeli są minimalnymi wersjami naprawiającymi konkretne wpisy, a nie zaleceniem, by używać dokładnie tych wersji.** Audyt badał Go 1.27.0. W dniu 2026-10-24 dostępna jest Go 1.27.3 (wydana 2026-10-19); nie ma jeszcze dowodu, że lokalna wersja została zaktualizowana lub że wykonano skan po naprawach. [Historia wydań Go](https://go.dev/doc/devel/release).

## Pełen rejestr 22 ustaleń audytu

B = błąd lub ograniczenie potwierdzone w kodzie; R = ryzyko bez odtworzonego skutku; M = modernizacja. Priorytet wynika z lokalnego zastosowania laboratoryjnego opisanego w audycie, **nie z CVSS**. Łącznie: 9 wysokich, 12 średnich, 1 niski.

| ID | Klasa | Priorytet | Co wykryto | Uzgodniony kierunek | Status |
| --- | --- | --- | --- | --- | --- |
| B01 | B | Wysoki | Stałe hasło PFX `changeit`, ekspozycja hasła i tryb pliku `0644` | Etap 3; funkcję PFX w kreatorze blokować do poprawki | Otwarte |
| B02 | B | Wysoki | Przestarzałe parametry kryptograficzne PKCS#12 (RC2/3DES/SHA-1) | Etap 3; nowy profil i test importu IIS/OpenSSL | Otwarte |
| B03 | B | Wysoki | Nadpisywanie plików, brak atomowego zapisu i kontroli uprawnień | Bezpieczne tworzenie, zapis pary i potwierdzanie nadpisania | Otwarte |
| B04 | B | Średni | Tworzenie nowego CA przed walidacją operacji, także przy `-uninstall` | Rozdzielenie odczytu CA od jego tworzenia | Otwarte |
| B05 | B | Wysoki | Usuwanie certyfikatów Windows wyłącznie na podstawie numeru seryjnego | Ścisła identyfikacja DER/odcisk SHA-256 oraz test kolizji serialu | Otwarte |
| B06 | B | Średni | Możliwy `panic` podczas obsługi CSR bez obsługiwanych SAN | Walidacja tożsamości CSR i regresja pustego SAN | Otwarte |
| B07 | B | Średni | Możliwość mylącej informacji o udanej instalacji zaufania | Jednoznaczne wyniki per magazyn i kontrola stanu | Otwarte |
| B08 | B | Wysoki | Dwie podatności w osiągalnej ścieżce IDNA/Unicode (`x/net`, `x/text`) | Aktualizacja zależności, testy błędnych nazw i ponowny govulncheck | Otwarte |
| B09 | B | Średni | Go 1.27.0 bez poprawki weryfikacji `crypto/x509` | Budowanie poprawionym Go 1.27.3+ oraz ponowny skan | Otwarte |
| R01 | R | Wysoki | Niezweryfikowane ACL kluczy i katalogu CA na Windows | Test ACL i ochrony CAROOT na obsługiwanych systemach | Otwarte |
| R02 | R | Wysoki | Nieatomowe tworzenie CA, dowiązania i równoległe uruchomienia | Zabezpieczenie inicjalizacji CA, ścieżek i stanów awaryjnych | Otwarte |
| R03 | R | Średni | Niepełna walidacja CA i ograniczeń dat certyfikatu | Kontrola CA, zgodności klucza i okresów ważności | Otwarte |
| R04 | R | Wysoki | Kopiowanie rozszerzeń CSR bez polityki wystawcy | Weryfikacja SAN/KU/EKU/OID i ograniczenie profilu HTTPS | Otwarte |
| R05 | R | Średni | Niebezpieczne wyznaczanie nazw ścieżek z danych SAN/URI | Oddzielenie identyfikatorów SAN od nazw plików | Otwarte |
| R06 | R | Średni | Niejasny zakres instalacji zaufania Windows (CurrentUser vs LocalMachine) | Jasne komunikaty i test konta docelowego | Otwarte |
| R07 | R | Średni | Częściowe operacje Linux/NSS i usuwanie legacy bez ścisłej identyfikacji | Dokładne dopasowanie i obsługa awarii kroków | Otwarte |
| R08 | R | Średni | Argumenty `keytool`, brak timeoutów i zakończenia przez `Fatal` | Bezpieczne procesy, sprzątanie i obsługa błędów | Otwarte |
| M01 | M | Wysoki | Brak testów jednostkowych i atrap magazynów zaufania | Izolowane testy regresji przed naprawą krytycznych ścieżek | Otwarte |
| M02 | M | Średni | Niepowtarzalne CI: stare akcje i ruchome wersje narzędzi | Ustalenie wersji i kontroli vet, test, vulncheck | Otwarte |
| M03 | M | Średni | Release bez pełnej dokumentacji, licencji i bramki testowej | Pakiet LICENSE/AUTHORS, informacje licencyjne zależności oraz testy przed publikacją | Otwarte |
| M04 | M | Średni | Niespójne wymagania Go i odziedziczona dokumentacja/nazewnictwo | Ustalenie wspieranego Go; rebranding w etapie 1 | Otwarte |
| M05 | M | Niski | Drobne ograniczenia DNS/seriali/API oraz macOS | Modernizacja po wyższych priorytetach | Otwarte |

Wszelkie zmiany stanu ustaleń muszą odsyłać do konkretnego wpisu w końcowym dzienniku i dowodów testowych. Nie należy oznaczać pozycji B01/B02 jako naprawionych tylko dlatego, że funkcja PFX została wyłączona w kreatorze — można wtedy co najwyżej odnotować ograniczenie ekspozycji.

## Uporządkowanie najbliższych prac (plan, NIE wykonanie)

Poniższa kolejność nie zmienia zatwierdzonej roadmapy. Rozdziela planowanie etapu 0 od implementacji etapów następnych.

1. Zamknąć i opublikować zweryfikowany audyt oraz ten rejestr; na `audit/stage0` nie zmieniać kodu aplikacji.
2. Ustalić wspieraną wersję Go i zgodny zestaw zależności. W osobnym PR zaktualizować kompilator/biblioteki, wykonać `go vet`, `go test`, kompilację Windows/Linux i `govulncheck`; dopiero wtedy ewentualnie zamykać GO-2026-5970, GO-2026-5026 oraz GO-2026-6809.
3. Zaplanować regresje i izolowane testy dla bezpiecznego zapisu, CA i magazynów zaufania (M01); nie przeprowadzać rzeczywistych operacji `-install` / `-uninstall` w testach jednostkowych.
4. W osobnych poprawkach usunąć niebezpieczne zachowania B03, B05 oraz B04, a następnie obsłużyć B06/B07 i uzasadnione R01–R08. Każdą zmianę testować i dokumentować.
5. Zrealizować ochronę PFX z B01/B02 zgodnie z etapem 3 `ROADMAP.md`; kreator etapu 2 do tej chwili nie udostępnia niezabezpieczonego eksportu PFX.
6. Po ukończeniu funkcji i testów wersji 1.0 przygotować publiczny opis **What QuestLab Cert adds to mkcert** oparty wyłącznie na zweryfikowanych wpisach dziennika.

## Podsumowanie do wydania 1.0

Ta sekcja jest celowo **niezamknięta**; zostanie opracowana na podstawie wpisów dziennika przed wydaniem 1.0.

- Liczba poprawek kodu wykonanych przez QuestLab: do wyliczenia z wpisów PR.
- Liczba **zweryfikowanych** usuniętych podatności GO/CVE/GHSA: do wyliczenia po ponownych skanach.
- Najważniejsze ulepszenia bezpieczeństwa i niezawodności: do udokumentowania.
- Nowe funkcje względem oryginalnego `mkcert`: do udokumentowania.
- Status Windows/Linux, testów HTTPS i ochrony kluczy: do potwierdzenia.
- Otwarte ograniczenia i ryzyka zaakceptowane w wydaniu: do wypisania.

---

## Dziennik wykonanych zmian — dopisuj nowe wpisy wyłącznie na końcu

### 001 — 2026-10-24 — poprawka enumeracji certyfikatów Windows

- **Kategoria:** poprawka techniczna / obsługa zasobów, bez przypisanego CVE/GO/GHSA.
- **Stan przed:** `go vet` zgłaszał `possible misuse of unsafe.Pointer` w `truststore_windows.go`.
- **Co zrobił QuestLab:** zastąpiono ręczną konwersję przez `syscall.CertEnumCertificatesInStore`; zachowano `CRYPT_E_NOT_FOUND`, dodano zwalnianie kontekstu na wcześniejszym wyjściu.
- **Kod i dowód:** [PR #1](https://github.com/mariuszsmolka-stack/qlcert/pull/1); commit `ebbca46`, merge `4ef25cb`.
- **Kontrole:** zgłoszono poprawne `go vet ./...`, `go test ./...`, build Windows i `git diff --check`. Nie było wówczas testów jednostkowych.
- **Ograniczenie:** nie przeprowadzono testów funkcjonalnych instalacji i usuwania certyfikatów w magazynie Windows.
- **Status:** scalono i zweryfikowano statycznie; nie ogłaszamy usunięcia konkretnego CVE.

### 002 — 2026-10-24 — plan rozwoju QuestLab Cert 1.0

- **Kategoria:** dokumentacja i organizacja prac.
- **Stan przed:** brak zatwierdzonego planu QuestLab Cert 1.0.
- **Co zrobił QuestLab:** przyjęto `ROADMAP.md` (197 linii): zastosowanie laboratoryjne HTTPS, etapy 0–6, EN domyślnie, PL przez `-lang=pl`, kreator, bezpieczeństwo PFX i scenariusz akceptacyjny HTTP → HTTPS.
- **Kod i dowód:** [PR #2](https://github.com/mariuszsmolka-stack/qlcert/pull/2); commit `23f8487`, merge `fcf0105`.
- **Kontrole:** przejrzano dokument przed scaleniem.
- **Status:** scalono; nie jest to usunięcie podatności.

<!-- Następny zweryfikowany i scalony zakres otrzymuje nagłówek ### 003 — RRRR-MM-DD — ... na KOŃCU dokumentu. Sam audyt, przed jego scaleniem, pozostaje wyłącznie w rejestrze ustaleń. -->

### 003 — 2026-10-24 — aktualizacja lokalnego środowiska Go do 1.27.3

- **Kategoria:** lokalne środowisko kompilacji / weryfikacja bezpieczeństwa; ustalenie B09.
- **Stan przed:** lokalne Go 1.27.0; `govulncheck` zgłaszał trzy podatności z osiągalnymi wywołaniami: GO-2026-6809, GO-2026-5970 oraz GO-2026-5026.
- **Co zrobił QuestLab:** zaktualizowano lokalne środowisko Go z wersji 1.27.0 do 1.27.3.
- **Kontrole i stan po:** ponowny `govulncheck` zgłasza dwie podatności z osiągalnymi wywołaniami. GO-2026-6809 nie jest już zgłaszana; GO-2026-5970 oraz GO-2026-5026 pozostają otwarte. Skan wykazał dodatkowo 49 zgłoszeń na poziomie modułów bez wskazanych osiągalnych wywołań; nie oznacza to 49 potwierdzonych możliwości zaatakowania aplikacji.
- **Zakres i pliki:** zmiana Go dotyczy na razie wyłącznie lokalnego środowiska. Nie zmieniono konfiguracji GitHub Actions ani zależności projektu. W tym wpisie aktualizowany jest wyłącznie `docs/DEVELOPMENT_HISTORY.md`.
- **Commit / PR:** nie wskazano dla aktualizacji lokalnego środowiska.
- **Status i ograniczenie:** GO-2026-6809 nie jest zgłaszana w ponownym skanie lokalnym. Nie uznajemy jej za usuniętą ze wszystkich przyszłych wydań programu; wymaga to osobnej weryfikacji środowiska kompilacji i wynikowych artefaktów. GO-2026-5970 i GO-2026-5026 pozostają otwarte.

### 004 — 2026-10-24 — aktualizacja zależności Go i końcowa kontrola bezpieczeństwa

- **Kategoria:** zależności / weryfikacja bezpieczeństwa; ustalenie B08.
- **Stan przed:** po aktualizacji lokalnego Go opisanej we wpisie 003 pozostawały dwie podatności z osiągalnymi wywołaniami (GO-2026-5970 i GO-2026-5026) oraz 49 dodatkowych zgłoszeń modułowych bez wskazanych osiągalnych wywołań.
- **Co zrobił QuestLab:** zaktualizowano `golang.org/x/net` do `v0.61.0` i `golang.org/x/text` do `v0.43.0`. Wymagania nowego `x/net` wymusiły aktualizację pośredniej zależności `golang.org/x/crypto` do `v0.58.0` oraz podniesienie minimalnej wersji Go w `go.mod` z `1.18` do `1.26.0`. Nie aktualizowano niezależnie biblioteki PKCS#12 ani `plist`.
- **Pliki i środowisko:** aktualizacja zależności obejmuje `go.mod` i `go.sum`; końcowe kontrole wykonano lokalnym Go `1.27.3` na gałęzi `fix/go-dependencies`. Podczas końcowej kontroli dopisano wyłącznie ten wpis w `docs/DEVELOPMENT_HISTORY.md`; nie zmieniano kodu aplikacji, audytu ani GitHub Actions.
- **Kontrole:** `go mod verify` — kod 0, `all modules verified`; `go vet ./...` — kod 0, bez ostrzeżeń; `go test ./...` — kod 0, `[no test files]`. Kompilacje Windows AMD64 i Linux AMD64 z `CGO_ENABLED=0` zakończyły się kodem 0. Linux sprawdzono przez kompilację krzyżową na Windows; pliki wynikowe zapisano w katalogu tymczasowym i nie uruchamiano aplikacji. Kontrole korzystały z `GOTOOLCHAIN=local` i `GOFLAGS=-mod=readonly`.
- **Ponowny skan:** `govulncheck v1.9.0`, uruchomiony przez `go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 -show verbose ./...`, zakończył się kodem 0. Wynik lokalnego skanu Windows: 0 podatności z osiągalnymi wywołaniami, 0 dodatkowych zgłoszeń na poziomie importowanych pakietów i 1 dodatkowe zgłoszenie modułowe bez wskazanego osiągalnego wywołania.
- **Podatności i ograniczenia:** GO-2026-5970 i GO-2026-5026 nie są wykrywane w ponownym skanowaniu lokalnym. Pozostaje GO-2026-5932 dotycząca nieutrzymywanego `golang.org/x/crypto/openpgp`; skaner nie wskazuje wywołania tej podatnej funkcjonalności przez aplikację i nie podaje wersji naprawiającej. GO-2026-6809 również nie jest zgłaszana przy lokalnym Go 1.27.3; zgodnie z wpisem 003 nie oznacza to potwierdzenia jej usunięcia ze wszystkich przyszłych wydań. Środowisko GitHub Actions i artefakty wydania wymagają osobnej weryfikacji.
- **PFX:** eksport nadal wymaga osobnych poprawek B01/B02 z etapu 3 roadmapy, w tym usunięcia stałego hasła `changeit` i poprawy parametrów ochrony PKCS#12. Aktualizacja zależności nie zamyka tych ustaleń.
- **Commit / PR:** nie wykonano commit ani push; nie wskazano PR dla tego wpisu.
- **Status:** wymagane końcowe kontrole lokalne zakończone powodzeniem. Brak plików testowych nie jest potwierdzeniem pokrycia funkcji testami; nie deklarujemy zakończenia etapu 0 ani weryfikacji wszystkich przyszłych wydań.

### 005 — 2026-10-24 — pierwszy zestaw testów jednostkowych

- **Kategoria:** testy bazowe / częściowa realizacja M01.
- **Stan przed:** brak plików testowych; `go test` zgłaszał `[no test files]`.
- **Co zrobił QuestLab:** dodano `cert_test.go` i `main_test.go`: 5 funkcji testowych oraz 29 przypadków table-driven dla `fileNames`, `randomSerialNumber`, `storeEnabled`, `getCAROOT` i `pathExists`. Testy używają `t.TempDir()` i `t.Setenv()`, bez równoległych zmian środowiska, generowania CA i operacji na magazynach zaufania.
- **Kontrole:** na Windows AMD64 z Go 1.27.3 wykonano `go test -v ./...` — wszystkie przypadki przeszły; `go test -race ./...` — poprawnie, bez wykrytych wyścigów; `go vet ./...` — bez ostrzeżeń. Wszystkie kontrole zwróciły kod 0. Poprawnie zakończyły się również `go mod verify`, kompilacje Windows AMD64 i Linux AMD64 oraz `git diff --check`.
- **Status i ograniczenia:** zestaw zatwierdzono. Testy wykonano na Windows; Linux sprawdzono wyłącznie przez kompilację krzyżową, bez uruchamiania testów na Linux. Jest to pierwszy zestaw bazowy, bez deklaracji pełnego pokrycia kodu ani zamknięcia M01. Nie zmieniono logiki aplikacji; numery seryjne nadal mogą wynosić zero.

### 006 — 2026-10-24 — ochrona zapisu PEM przed nadpisaniem

- **Kategoria:** bezpieczeństwo zapisu plików / częściowa realizacja B03.
- **Stan przed:** zapis PEM mógł nadpisać istniejący certyfikat lub klucz, a błąd drugiego zapisu pozostawić niekompletną parę.
- **Co zrobił QuestLab:** dodano wyłączne tworzenie plików (`O_EXCL`) i czytelny błąd odmowy nadpisania. Przy obsłużonym błędzie program wycofuje pliki utworzone przez bieżącą operację, sprawdzając ich tożsamość; nie usuwa ani nie modyfikuje plików istniejących przed operacją. Błędy sprzątania są dołączane do komunikatu. Ochrona obejmuje osobne pliki PEM, wspólny plik certyfikatu i klucza oraz certyfikat zapisany z CSR. Pliki: `cert.go`, `pemfile.go`, `pemfile_test.go`.
- **Kontrole:** `go test ./...` i `go test -race ./...` zakończone powodzeniem na Windows AMD64 z Go 1.27.3; `go vet ./...` bez ostrzeżeń. Testy obejmują odmowę nadpisania, zachowanie istniejących plików, wycofanie po błędzie drugiego zapisu i jednoczesne tworzenie tego samego pliku. Kompilacje Windows AMD64 i krzyżowa Linux AMD64 oraz `git diff --check` zakończyły się poprawnie. Wszystkie kontrole zwróciły kod 0; testów nie uruchamiano na Linux.
- **Status i ograniczenia:** poprawkę zatwierdzono; B03 pozostaje częściowo otwarte. Tworzenie CA i eksport PFX pozostają poza zakresem. Wycofywanie po obsłużonym błędzie nie zapewnia atomowej transakcji pary plików odpornej na nagłe przerwanie procesu.
