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

**Stan na 24.10.2026, po wpisie 014 w [historii rozwoju](docs/DEVELOPMENT_HISTORY.md).** Oznaczenia dotyczą implementacji i kontroli lokalnych, nie gotowości wydania 1.0.

| Obszar | Zrealizowano | Pozostaje do sprawdzenia lub wykonania |
| --- | --- | --- |
| Interfejs | Nazwa QuestLab Cert, pomoc i komunikaty EN/PL, domyślnie EN, `-lang=pl`, menu interaktywne, zachowane CLI | Dopracowanie kreatora dla wielu nazw SAN |
| HTTPS i PEM | Generowanie dla DNS/IP, lokalne CA, zapis PEM w katalogu użytkownika, odmowa nadpisania | Kontrola pozostałych ryzyk ochrony CA i zapisu |
| PFX | Eksport z hasłem podawanym bez echa, profil Modern2023, ochrona plików, obsługa w menu EN/PL | Import na rzeczywistym IIS; pełna weryfikacja środowisk docelowych |
| Instrukcje | Instrukcje Apache, Nginx i IIS w EN/PL, bez automatycznej ingerencji w konfigurację | Testy na działających serwerach i sprawdzenie zaufania klienta |
| Testy i zależności | Testy jednostkowe i regresyjne na Windows, `go vet`, build Windows i Linux AMD64, lokalny `govulncheck`: 0 zgłoszonych osiągalnych podatności | Testy wykonania na Linux, integracyjne, powtarzalne CI i artefakty wydania |
| Otwarte ustalenia audytu | Naprawiono m.in. B01, B02, B04; częściowo B03 | W szczególności B05 (dokładne dopasowanie przy usuwaniu zaufania Windows), inne niezamknięte ustalenia z audytu |

Pełny wykaz ustaleń i ograniczeń znajduje się w [TECHNICAL_AUDIT.md](docs/TECHNICAL_AUDIT.md); chronologia poprawek w [DEVELOPMENT_HISTORY.md](docs/DEVELOPMENT_HISTORY.md).

## Zakres funkcjonalny wersji 1.0

- Tworzenie lokalnego CA i czytelne wyjaśnienie jego roli.
- Generowanie certyfikatów dla nazw DNS i adresów IP przez kreator lub CLI.
- Wybór formatu PEM/PFX oraz bezpieczna obsługa kluczy prywatnych.
- Instrukcje podłączania HTTPS do IIS, Apache i Nginx.
- Sprawdzanie właściwości certyfikatu, objętych nim nazw i adresów oraz okresu ważności.
- Wyjaśnienie instalacji zaufania do lokalnego CA i jej skutków.
- Czytelne komunikaty oraz pomoc w językach PL/EN.

## Etapy rozwoju

Listy kontrolne pokazują stan realizacji według kodu i udokumentowanych kontroli. Testy na rzeczywistych serwerach pozostają odrębnym warunkiem wydania.

### Etap 0 — uporządkowanie i sprawdzenie istniejącego kodu

**Status:** audyt wykonany; wskazane problemy są naprawiane zgodnie z priorytetami.

**Cel:** ustalić sprawdzoną bazę techniczną do dalszego rozwoju.

**Zakres:** przegląd istniejących funkcji, ostrzeżeń narzędzi Go, obsługi zasobów Windows i ograniczeń bezpieczeństwa; sprawdzenie zachowania informacji licencyjnych. Przegląd zależności Go pod kątem aktualności, bezpieczeństwa i kompatybilności oraz ocena bibliotek kryptograficznych, szczególnie obsługi PKCS#12.

**Kryteria ukończenia:**

- [x] Poprawiono enumerację w `truststore_windows.go` bez wyciszania `go vet`; sprawdzono koniec enumeracji i zwalnianie kontekstów oraz wykonano `go vet`, `go test` i build Windows.
- [x] Spisano aktualne funkcje, zależności platformowe i problemy wymagające rozwiązania, w tym stałe hasło eksportu PKCS#12.
- [x] Udokumentowano przegląd zależności Go: ich aktualność, bezpieczeństwo i kompatybilność oraz ocenę bibliotek kryptograficznych, ze szczególnym uwzględnieniem obsługi PKCS#12.
- [x] Przejrzano pozostałą obsługę błędów i zasobów w istniejącym kodzie oraz określono braki testów; wynik `go test` bez plików testowych nie jest uznawany za pokrycie testami.
- [x] Potwierdzono zachowanie `LICENSE`, `AUTHORS` i nagłówków licencyjnych oraz zapisano wyniki przeglądu i listę otwartych prac.

### Etap 1 — identyfikacja QuestLab Cert i interfejs PL/EN

**Status:** ukończony w zakresie interfejsu EN/PL i menu (wpisy 008–009).

**Cel:** nadać programowi spójną nazwę QuestLab Cert i przygotować dwujęzyczną obsługę.

**Zakres:** identyfikacja programu, angielski jako język domyślny, obsługa `-lang=pl` i `-lang=en`, menu startowe, tłumaczenia pomocy i komunikatów oraz przygotowanie tekstów dla kreatora z zachowaniem CLI. Bez automatycznego wykrywania języka systemu i bez zapisywania preferencji językowych.

**Kryteria ukończenia:**

- [x] Pomoc i informacja o wersji identyfikują program jako QuestLab Cert oraz zachowują informację o pochodzeniu z mkcert.
- [x] Bez parametru `-lang` cała aplikacja używa EN; jawne `-lang=en` daje ten sam wybór języka, a `-lang=pl` wybiera PL. Każdy dostępny komunikat i tekst pomocy ma odpowiednik w obu językach.
- [x] Parametr `-lang` stosuje ten sam wybór języka w menu, kreatorze interaktywnym i tradycyjnym CLI; sprawdzono oba tryby dla PL oraz EN.
- [x] Uruchomienie bez argumentów otwiera menu interaktywne w EN. Uruchomienie z samym `-lang=pl` otwiera menu w PL, a z samym `-lang=en` — w EN.
- [x] Zmiana języka systemu nie wpływa na język aplikacji. Po uruchomieniu z `-lang=pl` kolejne uruchomienie bez tego parametru używa EN i nie odczytuje ani nie zapisuje preferencji językowych.
- [x] Nieprawidłowy kod, np. `-lang=de`, powoduje czytelny komunikat błędu wskazujący dozwolone wartości `en` i `pl`.
- [x] Te same scenariusze pomocy, poprawnego wywołania i błędnych argumentów sprawdzono dla domyślnego EN, jawnego `-lang=en` oraz `-lang=pl`; istniejące polecenia CLI nadal działają.

### Etap 2 — interaktywny kreator certyfikatów HTTPS

**Status:** w realizacji; podstawowy kreator PEM/PFX działa (wpisy 009–010, 014).

**Cel:** umożliwić początkującemu użytkownikowi przygotowanie certyfikatu HTTPS bez znajomości składni CLI.

**Zakres:** wybór pojedynczej nazwy DNS lub adresu IP, formatu PEM/PFX, katalogu zapisu, podsumowanie i potwierdzenie. Bezpieczny eksport PFX jest już dostępny; obsługa wielu nazw SAN w kreatorze pozostaje planowana.

**Kryteria ukończenia:**

- [x] Kreator w PL i EN prowadzi od podania DNS/IP do podsumowania i wygenerowania certyfikatu.
- [x] Kreator oferuje PEM lub PFX dopiero po wdrożeniu eksportu z hasłem użytkownika i profilem Modern2023, bez stałego `changeit`.
- [x] Sprawdzono pojedynczą nazwę DNS, IPv4, IPv6 i odrzucenie pustych lub błędnych danych.
- [ ] Dodano i przetestowano w kreatorze wiele nazw DNS/IP (SAN) dla jednego certyfikatu.
- [x] Użytkownik może anulować przed zapisem; istniejące pliki są chronione przed nadpisaniem (operacja jest odrzucana).
- [x] Certyfikat zawiera wskazaną pojedynczą nazwę lub adres, a dotychczasowe CLI nadal działa.

### Etap 3 — bezpieczny eksport PEM/PFX i lokalne CA

**Status:** częściowo wykonany; bezpieczny eksport PFX działa, ochrona CA i testy docelowe pozostają otwarte.

**Cel:** zapewnić bezpieczne przechowywanie i eksport materiału kryptograficznego oraz świadomą obsługę lokalnego CA.

**Zakres:** ochrona kluczy prywatnych, bezpieczne wprowadzanie haseł PFX, weryfikacja algorytmów ochrony danych PKCS#12, tworzenie i ponowne używanie CA oraz zgoda na instalację zaufania.

**Kryteria ukończenia:**

- [x] W testach lokalnych sprawdzono utworzenie i ponowne użycie CA oraz zgodność klucza i certyfikatu w eksportach PEM/PFX.
- [x] Usunięto stałe hasło `changeit`; PFX wymaga hasła użytkownika z ukrytym wprowadzaniem i potwierdzeniem, bez przekazywania w argumentach CLI lub logach.
- [x] Zweryfikowano lokalnie profil Modern2023, algorytmy ochrony PKCS#12 oraz odczyt wygenerowanego pakietu.
- [ ] Potwierdzono kompatybilność poprzez import PFX na rzeczywistym IIS.
- [ ] Sprawdzono ograniczenia dostępu do plików kluczy na Windows, następnie Linux; standardowy eksport certyfikatu HTTPS nigdy nie zawiera klucza prywatnego CA.
- [ ] Kreator wyjaśnia skutki instalacji CA i wymaga osobnej zgody. W CLI instalację zaufania uruchamia wyłącznie jawne polecenie użytkownika; odmowa w kreatorze nie zmienia magazynu.
- [x] Generowanie i eksport działają lokalnie bez instalowania zaufania; testy używają katalogów tymczasowych i nie zmieniają systemowych magazynów.

### Etap 4 — instrukcje konfiguracji HTTPS

**Status:** instrukcje gotowe; testy integracyjne na serwerach oczekują.

**Cel:** umożliwić samodzielne podłączenie wygenerowanego certyfikatu do serwera WWW w laboratorium.

**Zakres:** instrukcje PL/EN dla IIS, Apache i Nginx, dobór PEM/PFX, wskazanie plików oraz wyjaśnienie zaufania na komputerze klienta.

**Kryteria ukończenia:**

- [x] Dostępne są trzy instrukcje: IIS, Apache i Nginx, każda w PL oraz EN.
- [x] Każda instrukcja opisuje wymagane pliki, konfigurację HTTPS, weryfikację działania oraz rozróżnia certyfikat CA i jego klucz prywatny.
- [ ] Każdy scenariusz sprawdzono w odizolowanym laboratorium; udokumentowano użyte wersje serwera i wynik połączenia HTTPS.
- [x] Instrukcje wyjaśniają, że zaufanie do CA jest osobnym działaniem użytkownika i nie powstaje przez samo wygenerowanie certyfikatu.

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
