# Przewodnik dla osób wnoszących wkład

Dziękujemy za zainteresowanie wniesieniem wkładu do projektu Kubernetes Demo Application (kuard)! Oto kilka wskazówek, które pomogą Ci rozpocząć.

## Zgłaszanie błędów

Przed zgłoszeniem nowego problemu:

1. Sprawdź, czy podobny problem nie został już zgłoszony w sekcji [Issues](https://github.com/yourusername/kuard/issues).
2. Jeśli nie znaleziono podobnego problemu, utwórz nowy, zawierający:
   - Krótki i opisowy tytuł
   - Szczegółowy opis problemu
   - Kroki do odtworzenia problemu
   - Oczekiwane i rzeczywiste zachowanie
   - Informacje o środowisku (system operacyjny, wersja Go, itp.)
   - Zrzuty ekranu (jeśli dotyczy)

## Propozycje funkcji

Mamy otwarty umysł na nowe funkcje! Aby zaproponować nową funkcję:

1. Sprawdź, czy podobna funkcja nie została już zaproponowana lub wdrożona.
2. Otwórz nowy problem z prefiksem [Feature Request] w tytule.
3. Opisz proponowaną funkcję, jej przypadki użycia i korzyści dla projektu.

## Pull Requesty

Zapraszamy do przesyłania pull requestów! Oto proces:

1. Utwórz fork tego repozytorium
2. Utwórz nową gałąź dla swojej funkcji lub poprawki błędu
3. Dokonaj swoich zmian
4. Upewnij się, że wszystkie testy przechodzą
5. Wyślij pull request do gałęzi głównej

### Konwencje kodowania

- Kod Go musi być sformatowany przy użyciu `go fmt`.
- Wszystkie publiczne funkcje i typy muszą być udokumentowane.
- Dodaj testy dla nowych funkcji lub poprawek błędów.
- Upewnij się, że Twój kod przechodzi `go vet`.

### Wytyczne dla commit messages

- Używaj imperatywu („Add feature
