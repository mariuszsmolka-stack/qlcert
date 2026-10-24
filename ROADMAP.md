# QuestLab Cert — plan rozwoju do wersji 1.0

## Cel projektu

QuestLab Cert to proste narzędzie do generowania i wdrażania certyfikatów X.509 do HTTPS w lokalnych laboratoriach informatycznych. Ma ułatwiać uruchamianie usług WWW oraz wyjaśniać zasady korzystania z lokalnego urzędu certyfikacji (CA).

Projekt jest rozwijany przez QuestLab na podstawie otwartego projektu mkcert. Zachowuje oryginalną licencję BSD-3-Clause, informacje o autorach i wymagane informacje licencyjne, dostępne w plikach [LICENSE](LICENSE) oraz [AUTHORS](AUTHORS).

Czytelny interfejs, bezpieczne ustawienia i sprawdzona dokumentacja mają również prezentować jakość i możliwości techniczne QuestLab.

## Użytkownicy

- Początkujący administratorzy.
- Nauczyciele informatyki.
- Uczniowie i uczestnicy szkoleń.
- Osoby uruchamiające własne usługi WWW w laboratoriach.

## Założenia i zakres

- Prosty, interaktywny kreator oraz zachowanie pracy z CLI.
- Interfejs, komunikaty i pomoc w językach polskim oraz angielskim (PL/EN).
- Domyślnym językiem całej aplikacji jest angielski (EN). Polski wybiera się parametrem `-lang=pl`; dopuszczalne jest także jawne `-lang=en`.
- Parametr `-lang` obowiązuje zarówno w kreatorze interaktywnym, jak i w tradycyjnym CLI.
- Uruchomienie bez argumentów otwiera menu interaktywne w języku angielskim.
- Program nie wykrywa automatycznie języka systemu i nie zapisuje preferencji językowych; wybór obowiązuje tylko dla bieżącego uruchomienia.
- Nieprawidłowy kod języka powoduje czytelny komunikat błędu.
- Windows jako pierwsza platforma wdrożenia i weryfikacji; Linux jako kolejna.
- Brak konieczności instalowania Dockera, bazy danych lub stałego serwera CA.
- Możliwość pracy offline po przygotowaniu programu; lokalne generowanie certyfikatów nie wymaga publicznego CA.
- Certyfikaty X.509 do HTTPS; certyfikaty SSH pozostają poza zakresem.
- Zachowanie istniejących funkcji mkcert, o ile nie kolidują z zakresem projektu i wymaganiami bezpieczeństwa.
- Wdrożenie HTTPS poprzez instrukcje dla IIS, Apache i Nginx; bez automatycznej konfiguracji tych serwerów.

## Stan obecny

Poniższe możliwości wynikają z istniejącego kodu odziedziczonego po mkcert. Ich obecność nie oznacza ukończenia planowanych etapów ani pełnej weryfikacji wydania QuestLab Cert 1.0.

| Obszar | Aktualne możliwości i ograniczenia |
| --- | --- |
| Obsługa programu | Dostępne CLI; pomoc i komunikaty nadal odwołują się do mkcert. Brak interaktywnego kreatora i pełnego interfejsu PL/EN. |
| Lokalne CA | Tworzenie i odczyt lokalnego CA; istnieją polecenia instalowania i usuwania zaufania. Docelowe zasady zgody i ochrony kluczy wymagają przeglądu. |
| Certyfikaty HTTPS | Generowanie certyfikatów dla nazw DNS i adresów IP. |
| Eksport | Dostępne PEM i PKCS#12 (`.p12`, format znany także jako PFX). Obecny eksport PKCS#12 używa stałego hasła `changeit`; nie spełnia to wymagań bezpieczeństwa wersji 1.0. |
| Platformy | Istnieje kod obsługi Windows i Linux. Priorytet rozwoju i weryfikacji QuestLab Cert to Windows, następnie Linux. |
| Enumeracja Windows | Wykonana poprawka `truststore_windows.go`: użycie `syscall.CertEnumCertificatesInStore()`, zachowanie obsługi końca enumeracji i zwalnianie kontekstu przy wcześniejszym wyjściu. |
| Weryfikacja poprawki Windows | `go vet ./...` bez ostrzeżeń, `go test ./...` zakończone powodzeniem z informacją o braku plików testowych oraz poprawny build Windows. Nie instalowano ani nie usuwano certyfikatów. |

Interaktywny kreator, pełna obsługa PL/EN, bezpieczny eksport PFX bez stałego hasła, instrukcje QuestLab dla serwerów HTTPS i diagnostyka certyfikatów są planowane. Nie należy przedstawiać ich jako gotowych funkcji.

## Planowane funkcjonalności

- Tworzenie lokalnego CA i czytelne wyjaśnienie jego roli.
- Generowanie certyfikatów dla nazw DNS i adresów IP przez kreator lub CLI.
- Wybór formatu PEM/PFX oraz bezpieczna obsługa kluczy prywatnych.
- Instrukcje podłączania HTTPS do IIS, Apache i Nginx.
- Sprawdzanie właściwości certyfikatu, objętych nim nazw i adresów oraz okresu ważności.
- Wyjaśnienie instalacji zaufania do lokalnego CA i jej skutków.
- Czytelne komunikaty oraz pomoc w językach PL/EN.

## Etapy rozwoju

Listy kontrolne określają mierzalne kryteria ukończenia. Zaznaczono wyłącznie wykonaną poprawkę Windows; pozostałe punkty są pracami do wykonania.

### Etap 0 — uporządkowanie i sprawdzenie istniejącego kodu

**Status:** w toku; etap nie jest ukończony.

**Cel:** ustalić sprawdzoną bazę techniczną do dalszego rozwoju.

**Zakres:** przegląd istniejących funkcji, ostrzeżeń narzędzi Go, obsługi zasobów Windows i ograniczeń bezpieczeństwa; sprawdzenie zachowania informacji licencyjnych. Przegląd zależności Go pod kątem aktualności, bezpieczeństwa i kompatybilności oraz ocena bibliotek kryptograficznych, szczególnie obsługi PKCS#12.

**Kryteria ukończenia:**

- [x] Poprawiono enumerację w `truststore_windows.go` bez wyciszania `go vet`; sprawdzono koniec enumeracji i zwalnianie kontekstów oraz wykonano `go vet`, `go test` i build Windows.
- [ ] Spisano aktualne funkcje, zależności platformowe i problemy wymagające rozwiązania, w tym stałe hasło eksportu PKCS#12.
- [ ] Udokumentowano przegląd zależności Go: ich aktualność, bezpieczeństwo i kompatybilność oraz ocenę bibliotek kryptograficznych, ze szczególnym uwzględnieniem obsługi PKCS#12.
- [ ] Przejrzano pozostałą obsługę błędów i zasobów w istniejącym kodzie oraz określono braki testów; wynik `go test` bez plików testowych nie jest uznawany za pokrycie testami.
- [ ] Potwierdzono zachowanie `LICENSE`, `AUTHORS` i nagłówków licencyjnych oraz zapisano wyniki przeglądu i listę otwartych prac.

### Etap 1 — identyfikacja QuestLab Cert i interfejs PL/EN

**Status:** planowany.

**Cel:** nadać programowi spójną nazwę QuestLab Cert i przygotować dwujęzyczną obsługę.

**Zakres:** identyfikacja programu, angielski jako język domyślny, obsługa `-lang=pl` i `-lang=en`, menu startowe, tłumaczenia pomocy i komunikatów oraz przygotowanie tekstów dla kreatora z zachowaniem CLI. Bez automatycznego wykrywania języka systemu i bez zapisywania preferencji językowych.

**Kryteria ukończenia:**

- [ ] Pomoc i informacja o wersji identyfikują program jako QuestLab Cert oraz zachowują informację o pochodzeniu z mkcert.
- [ ] Bez parametru `-lang` cała aplikacja używa EN; jawne `-lang=en` daje ten sam wybór języka, a `-lang=pl` wybiera PL. Każdy dostępny komunikat i tekst pomocy ma odpowiednik w obu językach.
- [ ] Parametr `-lang` stosuje ten sam wybór języka w menu, kreatorze interaktywnym i tradycyjnym CLI; sprawdzono oba tryby dla PL oraz EN.
- [ ] Uruchomienie bez argumentów otwiera menu interaktywne w EN. Uruchomienie z samym `-lang=pl` otwiera menu w PL, a z samym `-lang=en` — w EN.
- [ ] Zmiana języka systemu nie wpływa na język aplikacji. Po uruchomieniu z `-lang=pl` kolejne uruchomienie bez tego parametru używa EN i nie odczytuje ani nie zapisuje preferencji językowych.
- [ ] Nieprawidłowy kod, np. `-lang=de`, powoduje czytelny komunikat błędu wskazujący dozwolone wartości `en` i `pl`.
- [ ] Te same scenariusze pomocy, poprawnego wywołania i błędnych argumentów sprawdzono dla domyślnego EN, jawnego `-lang=en` oraz `-lang=pl`; istniejące polecenia CLI nadal działają.

### Etap 2 — interaktywny kreator certyfikatów HTTPS

**Status:** planowany.

**Cel:** umożliwić początkującemu użytkownikowi przygotowanie certyfikatu HTTPS bez znajomości składni CLI.

**Zakres:** zebranie nazw DNS i adresów IP, wybór miejsca zapisu, podsumowanie oraz potwierdzenie generowania. Kreator może początkowo obsługiwać wyłącznie PEM. Wybór PEM/PFX i eksport PFX będą dostępne dopiero po zaimplementowaniu bezpiecznego mechanizmu z etapu 3. Nie wolno udostępniać użytkownikowi eksportu ze stałym hasłem `changeit`. Docelowy bezpieczny eksport i obsługa CA są warunkiem ukończenia etapu 3.

**Kryteria ukończenia:**

- [ ] Kreator w PL i EN prowadzi od podania DNS/IP do podsumowania i wygenerowania certyfikatu.
- [ ] Do czasu wdrożenia bezpiecznego mechanizmu z etapu 3 kreator udostępnia wyłącznie PEM; sprawdzono, że użytkownik nie może uruchomić eksportu ze stałym hasłem `changeit`.
- [ ] Sprawdzono poprawną nazwę DNS, adres IPv4, adres IPv6, wiele wpisów oraz odrzucenie pustych lub błędnych danych.
- [ ] Użytkownik może anulować przed zapisem; istniejące pliki wymagają potwierdzenia przed nadpisaniem.
- [ ] Certyfikat zawiera wskazane nazwy i adresy, a równoważne wywołanie CLI nadal działa.

### Etap 3 — bezpieczny eksport PEM/PFX i lokalne CA

**Status:** planowany.

**Cel:** zapewnić bezpieczne przechowywanie i eksport materiału kryptograficznego oraz świadomą obsługę lokalnego CA.

**Zakres:** ochrona kluczy prywatnych, bezpieczne wprowadzanie haseł PFX, weryfikacja algorytmów ochrony danych PKCS#12, tworzenie i ponowne używanie CA oraz zgoda na instalację zaufania.

**Kryteria ukończenia:**

- [ ] Sprawdzono utworzenie i ponowne użycie CA oraz zgodność klucza i certyfikatu w eksportach PEM/PFX.
- [ ] Usunięto stałe hasło `changeit`; PFX wymaga niepustego hasła określonego przez użytkownika. Hasło nie może być przekazywane jawnym argumentem CLI ani trafiać do historii poleceń lub logów. Wpisywanie hasła jest niewidoczne; hasło nie jest zapisywane w pomocy ani kodzie.
- [ ] Zweryfikowano i udokumentowano bezpieczeństwo oraz kompatybilność algorytmów ochrony danych PKCS#12 używanych przez mechanizm eksportu PFX.
- [ ] Sprawdzono ograniczenia dostępu do plików kluczy na Windows, następnie Linux; standardowy eksport certyfikatu HTTPS nigdy nie zawiera klucza prywatnego CA.
- [ ] Kreator wyjaśnia skutki instalacji CA i wymaga osobnej zgody. W CLI instalację zaufania uruchamia wyłącznie jawne polecenie użytkownika; odmowa w kreatorze nie zmienia magazynu.
- [ ] Generowanie i eksport działają offline bez instalowania zaufania; testy używają plików tymczasowych i atrap operacji magazynu.

### Etap 4 — instrukcje konfiguracji HTTPS

**Status:** planowany.

**Cel:** umożliwić samodzielne podłączenie wygenerowanego certyfikatu do serwera WWW w laboratorium.

**Zakres:** instrukcje PL/EN dla IIS, Apache i Nginx, dobór PEM/PFX, wskazanie plików oraz wyjaśnienie zaufania na komputerze klienta.

**Kryteria ukończenia:**

- [ ] Dostępne są trzy instrukcje: IIS, Apache i Nginx, każda w PL oraz EN.
- [ ] Każda instrukcja opisuje wymagane pliki, konfigurację HTTPS, weryfikację działania oraz rozróżnia certyfikat CA i jego klucz prywatny.
- [ ] Każdy scenariusz sprawdzono w odizolowanym laboratorium; udokumentowano użyte wersje serwera i wynik połączenia HTTPS.
- [ ] Instrukcje wyjaśniają, że zaufanie do CA jest osobnym działaniem użytkownika i nie powstaje przez samo wygenerowanie certyfikatu.

### Etap 5 — diagnostyka certyfikatów

**Status:** planowany.

**Cel:** ułatwić ocenę, do czego służy certyfikat i czy jego okres ważności pozwala na użycie.

**Zakres:** odczyt właściwości certyfikatu, wystawcy, nazw DNS i adresów IP oraz dat ważności; komunikaty PL/EN.

**Kryteria ukończenia:**

- [ ] Diagnostyka pokazuje wystawcę, podmiot, nazwy DNS/IP i początek oraz koniec ważności.
- [ ] Testy obejmują certyfikat aktualnie ważny, wygasły, jeszcze nieważny i niepoprawny plik.
- [ ] Wynik rozróżnia poprawny okres ważności od zaufania klienta do CA; nie przedstawia lokalnego certyfikatu jako publicznie zaufanego.
- [ ] Te same wyniki i błędy mają czytelne komunikaty PL/EN, a diagnostyka nie modyfikuje certyfikatów ani magazynów zaufania.

### Etap 6 — testy, dokumentacja i wydanie 1.0

**Status:** planowany.

**Cel:** wydać sprawdzoną, udokumentowaną wersję QuestLab Cert 1.0.

**Zakres:** testy kreatora i CLI, kontrola bezpieczeństwa, dokumentacja użytkownika oraz przygotowanie wydania; weryfikacja Windows przed Linux.

**Kryteria ukończenia:**

- [ ] Zakończono kryteria etapów 0–5 i zapisano wyniki testów dla Windows; następnie sprawdzono Linux i jednoznacznie opisano status jego wsparcia w wydaniu.
- [ ] `go vet ./...`, `go test ./...` i build kończą się powodzeniem na platformach deklarowanych jako obsługiwane; testy obejmują krytyczne scenariusze, a nie tylko kompilację.
- [ ] Sprawdzono PL/EN, kreator, CLI, pracę offline, eksport i diagnostykę bez modyfikowania rzeczywistych magazynów zaufania.
- [ ] Dokumentacja opisuje uruchomienie, lokalne CA, ochronę kluczy, zgodę na zaufanie, konfigurację HTTPS oraz ograniczenia laboratoryjne.
- [ ] Początkujący administrator, korzystając z programu i dokumentacji QuestLab Cert, potrafi skonfigurować HTTPS dla wcześniej działającej strony HTTP w laboratorium oraz poprawnie ustanowić zaufanie klienta do lokalnego CA. Scenariusz przetestowano i udokumentowano.
- [ ] Wydanie zawiera program, dokumentację, oryginalną licencję i informacje o autorach; nie zawiera kluczy prywatnych, haseł ani roboczych CA.

## Wymagania bezpieczeństwa

Poniższe zasady obowiązują we wszystkich etapach i są warunkami wydania 1.0, a nie deklaracją, że obecny kod już je wszystkie spełnia.

- **Klucz prywatny CA:** pozostaje pod kontrolą właściciela laboratorium, z ograniczonym dostępem. Nie trafia do repozytorium, logów ani pakietów przekazywanych użytkownikom usług. Do ustanowienia zaufania przekazuje się certyfikat CA, nigdy jego klucz prywatny.
- **Klucze certyfikatów HTTPS:** eksport i zapis chronią klucz przed nieuprawnionym odczytem oraz przypadkowym nadpisaniem.
- **Hasła PFX:** brak stałych lub wspólnych haseł; użytkownik określa hasło eksportu. Program nie ujawnia go w komunikatach ani logach.
- **Zgoda na zaufanie:** samo tworzenie CA lub certyfikatu nie instaluje CA w magazynie. Instalacja wymaga świadomej zgody w kreatorze albo jawnego polecenia CLI.
- **Lokalne i publiczne zaufanie:** certyfikat lokalnego CA nie jest automatycznie zaufany przez inne komputery ani przeglądarki na innych urządzeniach. Poprawne DNS/IP i daty ważności nie zastępują zaufania do wystawcy.
- **Testy:** korzystają z tymczasowych katalogów, testowych CA i atrap operacji magazynu. Testy integracyjne prowadzi się w odizolowanych środowiskach z jednorazowymi magazynami; nie ingerują w rzeczywiste magazyny zaufania użytkownika lub hosta.
- **Zastosowanie:** narzędzie jest przeznaczone do środowisk laboratoryjnych i edukacyjnych. Nie stanowi rozwiązania do zarządzania produkcyjną PKI ani publicznymi certyfikatami.

## Poza zakresem wersji 1.0

- Certyfikaty SSH.
- Pełny system PKI.
- Serwer ACME.
- Konta użytkowników i baza danych.
- Automatyczne wdrażanie na urządzeniach sieciowych.
- Rozbudowany monitoring i ewidencja infrastruktury.
