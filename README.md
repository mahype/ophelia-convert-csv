# CsvWatcher

Kleines Tray-Tool, das einen Ordner überwacht und neue DATEV-CSV-Dateien automatisch
umsortiert:

- **Zeile 1** bleibt unverändert (EXTF-Metadaten).
- **Zeile 2** (Kopfzeile) wird komplett durch den 254-spaltigen DATEV-Debitoren-Header
  (`transform.TargetHeader`) ersetzt — die Daten stehen bereits in den richtigen Spalten,
  nur die Beschriftungen im Original-Export passen nicht zum DATEV-Import.
- **Zeile 3+** (Datenzeilen) bleiben inhaltlich unverändert, werden aber bei Bedarf rechts
  mit leeren Feldern auf 254 Spalten aufgefüllt.

Das Original landet danach im Unterordner `verarbeitet/`, die umgewandelte Datei nimmt
den Namen der Originaldatei im überwachten Ordner ein.

## Für Endnutzer

### macOS

1. `CsvWatcher.app` einmalig irgendwohin kopieren (z. B. nach `/Applications`).
2. Doppelklick → ein neues Icon erscheint oben rechts in der Menüleiste.
3. Beim ersten Start öffnet sich ein Ordner-Auswahl-Dialog. Den gewünschten Ordner
   auswählen. Ab jetzt wird jede CSV, die du dort reinlegst, automatisch umgewandelt.
4. Optional bei Login automatisch starten:
   *Systemeinstellungen → Allgemein → Anmeldeobjekte → „+" → CsvWatcher.app wählen*.

### Windows

1. `CsvWatcher.exe` einmalig irgendwohin kopieren (z. B. `C:\Tools\CsvWatcher\`).
2. Doppelklick → ein neues Icon erscheint im Tray (unten rechts neben der Uhr).
3. Beim ersten Start öffnet sich ein Ordner-Auswahl-Dialog. Ordner wählen, fertig.
4. Optional bei Anmeldung automatisch starten: Verknüpfung der `CsvWatcher.exe` in
   den Autostart-Ordner legen (`Win+R` → `shell:startup` → rechte Maustaste → „Neu → Verknüpfung").

### Tray-Menü

| Eintrag | Funktion |
|---|---|
| Überwachter Ordner: … | Zeigt den aktiven Pfad (nur Info). |
| Ordner ändern… | Öffnet den Ordner-Picker und wechselt auf einen anderen Ordner. |
| Ordner öffnen | Öffnet den überwachten Ordner im Finder / Explorer. |
| Datei manuell umwandeln… | Einmalige Umwandlung einer Datei außerhalb des Ordners (Ergebnis landet mit Suffix `_konvertiert` daneben). |
| Watcher pausieren / fortsetzen | Hält den Automatismus an bzw. nimmt ihn wieder auf. |
| Beenden | Tool schließen. |

### Speicherorte

- Einstellungen: `~/Library/Application Support/CsvWatcher/config.json` (Mac) bzw.
  `%APPDATA%\CsvWatcher\config.json` (Windows).
- Bei Fehlschlag: eine Datei `<name>.csv.error.log` wird neben dem Original angelegt.
  Das Original bleibt in solchen Fällen liegen (nichts geht verloren).

## Für Entwickler

### Projekt-Layout

```
cmd/csv-watcher/    Entrypoint + Tray-UI
internal/transform/ Kern-Transformation (3 Schritte) + Unit-Tests
internal/fileio/    CP1252-Read/Write, Move nach verarbeitet/
internal/watcher/   fsnotify-Wrapper, Debounce, Self-Event-Filter
internal/config/    JSON-Persistenz
build/              Build-Scripts + Icon-Generator
```

### Bauen

Voraussetzung: Go 1.22+ mit aktiviertem CGo (wegen systray / dialog).

```bash
# macOS (arm64/amd64 – je nach Host-Architektur):
./build/build-mac.sh
# Ergebnis: dist/mac/CsvWatcher.app

# Windows-EXE, cross-compiled vom Mac (ohne MinGW, ohne Docker):
./build/build-win.sh
# Ergebnis: dist/win/CsvWatcher.exe

# Alternativ direkt auf einem Windows-Rechner:
build\build-win.bat
```

Der Cross-Compile funktioniert, weil alle Windows-Abhängigkeiten (`tadvi/systray`,
`sqweek/dialog`, `beeep`) Pure-Go-Codepfade für das Windows-Target haben. Keine
CGo-Toolchain nötig.

### Tests

```bash
go test ./...
```

Die `internal/fileio`-Suite nutzt die Original-Fixture
`original_EXTF_Div-Adressen_DEMOOPH_all_2026-03-25_05-13.csv` als End-to-End-Probe.

### Icons neu erzeugen

```bash
go run build/gen-icons.go
```

Erzeugt `cmd/csv-watcher/icon.png` (macOS-Tray) und `cmd/csv-watcher/icon.ico` (Windows-Tray),
die per `//go:embed` in das Binary eingebettet werden.
