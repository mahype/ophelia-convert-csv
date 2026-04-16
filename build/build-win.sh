#!/usr/bin/env bash
# Cross-Compile CsvWatcher.exe vom Mac (oder Linux) aus.
# Funktioniert ohne MinGW, weil die Windows-Pfade der Abhängigkeiten pure Go sind.
# Aufruf aus dem Repo-Root: ./build/build-win.sh

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

OUT_DIR="dist/win"
OUT="$OUT_DIR/CsvWatcher.exe"

mkdir -p "$OUT_DIR"
rm -f "$OUT"

echo "→ Kompiliere für Windows amd64 (ohne CGo) …"
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
  go build -ldflags="-s -w -H windowsgui" -o "$OUT" ./cmd/csv-watcher

echo "✓ Fertig: $OUT ($(du -h "$OUT" | awk '{print $1}'))"
echo "  Diese Datei auf einen Windows-Rechner kopieren und doppelklicken."
