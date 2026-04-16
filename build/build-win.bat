@echo off
REM Baut CsvWatcher.exe fuer Windows 64-bit.
REM Auf einem Windows-Rechner im Repo-Root ausfuehren.

setlocal
pushd "%~dp0.."

if not exist dist\win mkdir dist\win

echo Kompiliere CsvWatcher.exe ...
set CGO_ENABLED=1
go build -o dist\win\CsvWatcher.exe -ldflags="-s -w -H windowsgui" .\cmd\csv-watcher

if errorlevel 1 (
    echo FEHLER beim Build.
    popd
    exit /b 1
)

echo.
echo Fertig: dist\win\CsvWatcher.exe
echo   Zum Starten: Doppelklick auf die exe.
popd
endlocal
