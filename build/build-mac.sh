#!/usr/bin/env bash
# Baut CsvWatcher.app für den aktuellen Mac (arm64 oder amd64).
# Aufruf aus dem Repo-Root: ./build/build-mac.sh

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

OUT_DIR="dist/mac"
APP="$OUT_DIR/CsvWatcher.app"
BIN="$APP/Contents/MacOS/CsvWatcher"

rm -rf "$OUT_DIR"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"

echo "→ Binary wird kompiliert …"
CGO_ENABLED=1 go build -o "$BIN" -ldflags="-s -w" ./cmd/csv-watcher

cp "cmd/csv-watcher/icon.png" "$APP/Contents/Resources/AppIcon.png"

cat >"$APP/Contents/Info.plist" <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key><string>CsvWatcher</string>
  <key>CFBundleDisplayName</key><string>CsvWatcher</string>
  <key>CFBundleIdentifier</key><string>com.example.csvwatcher</string>
  <key>CFBundleVersion</key><string>1.0</string>
  <key>CFBundleShortVersionString</key><string>1.0</string>
  <key>CFBundleExecutable</key><string>CsvWatcher</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>LSMinimumSystemVersion</key><string>11.0</string>
  <key>LSUIElement</key><true/>
</dict>
</plist>
PLIST

echo "✓ Fertig: $APP"
echo "  Zum Starten: open $APP"
