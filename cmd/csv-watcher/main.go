package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"

	"csvwatcher/internal/config"
	"csvwatcher/internal/fileio"
	"csvwatcher/internal/transform"
	"csvwatcher/internal/watcher"
)

//go:embed icon.png
var iconPNG []byte

//go:embed icon.ico
var iconICO []byte

var (
	w           *watcher.Watcher
	mFolderInfo *systray.MenuItem
	mOpenFolder *systray.MenuItem
	mPause      *systray.MenuItem
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	if runtime.GOOS == "windows" {
		systray.SetIcon(iconICO)
	} else {
		systray.SetIcon(iconPNG)
	}
	systray.SetTooltip("CSV-Watcher – DATEV-Spaltenumsortierung")

	cfg, err := config.Load()
	if err != nil {
		log.Printf("load config: %v", err)
	}

	mFolderInfo = systray.AddMenuItem("Überwachter Ordner: (nicht gesetzt)", "")
	mFolderInfo.Disable()
	mChangeFolder := systray.AddMenuItem("Ordner ändern…", "Neuen Ordner zum Überwachen wählen")
	mOpenFolder = systray.AddMenuItem("Ordner öffnen", "Ordner im Finder/Explorer anzeigen")
	systray.AddSeparator()
	mManual := systray.AddMenuItem("Datei manuell umwandeln…", "Eine einzelne CSV umwandeln")
	systray.AddSeparator()
	mPause = systray.AddMenuItem("Watcher pausieren", "Automatisches Umwandeln anhalten")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Beenden", "CSV-Watcher schließen")

	w = watcher.New(cfg.WatchFolder, processFile)

	if cfg.WatchFolder != "" {
		if err := w.Start(); err != nil {
			log.Printf("start watcher: %v", err)
			beeep.Alert("CsvWatcher", fmt.Sprintf("Ordner kann nicht überwacht werden: %v", err), "")
		}
	}
	if cfg.Paused {
		w.SetPaused(true)
		mPause.SetTitle("Watcher fortsetzen")
	}
	refreshFolderMenu(cfg.WatchFolder)

	go handleMenus(mChangeFolder, mManual, mQuit)

	if cfg.WatchFolder == "" {
		go func() {
			_ = beeep.Notify("CsvWatcher", "Bitte einen Ordner zum Überwachen auswählen.", "")
			chooseFolder()
		}()
	}
}

func onExit() {
	if w != nil {
		w.Stop()
	}
}

func handleMenus(mChangeFolder, mManual, mQuit *systray.MenuItem) {
	for {
		select {
		case <-mChangeFolder.ClickedCh:
			chooseFolder()
		case <-mOpenFolder.ClickedCh:
			if folder := w.Folder(); folder != "" {
				openInFileExplorer(folder)
			}
		case <-mManual.ClickedCh:
			manualConvert()
		case <-mPause.ClickedCh:
			togglePause()
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func chooseFolder() {
	picked, err := dialog.Directory().Title("Ordner zum Überwachen auswählen").Browse()
	if err != nil {
		return
	}
	if err := w.SetFolder(picked); err != nil {
		_ = beeep.Alert("CsvWatcher", fmt.Sprintf("Ordner konnte nicht überwacht werden: %v", err), "")
		return
	}
	cfg, _ := config.Load()
	cfg.WatchFolder = picked
	if err := config.Save(cfg); err != nil {
		log.Printf("save config: %v", err)
	}
	refreshFolderMenu(picked)
	_ = beeep.Notify("CsvWatcher", "Ordner aktiv: "+picked, "")
}

func manualConvert() {
	picked, err := dialog.File().
		Title("CSV-Datei zum Umwandeln auswählen").
		Filter("CSV-Dateien", "csv", "CSV").
		Load()
	if err != nil {
		return
	}
	processFile(picked)
}

func togglePause() {
	newState := !w.IsPaused()
	w.SetPaused(newState)
	if newState {
		mPause.SetTitle("Watcher fortsetzen")
		_ = beeep.Notify("CsvWatcher", "Watcher pausiert.", "")
	} else {
		mPause.SetTitle("Watcher pausieren")
		_ = beeep.Notify("CsvWatcher", "Watcher aktiv.", "")
	}
	cfg, _ := config.Load()
	cfg.Paused = newState
	_ = config.Save(cfg)
}

func refreshFolderMenu(path string) {
	if path == "" {
		mFolderInfo.SetTitle("Überwachter Ordner: (nicht gesetzt)")
		mOpenFolder.Disable()
		return
	}
	mFolderInfo.SetTitle("Überwachter Ordner: " + path)
	mOpenFolder.Enable()
}

func processFile(path string) {
	rows, err := fileio.ReadCSV(path)
	if err != nil {
		fileio.WriteErrorSidecar(path, err)
		_ = beeep.Alert("CsvWatcher", fmt.Sprintf("Fehler beim Lesen von %s: %v", filepath.Base(path), err), "")
		return
	}
	transform.Apply(rows)

	folder := w.Folder()
	inWatched := folder != "" && filepath.Dir(path) == folder

	if inWatched {
		archivePath, err := fileio.MoveToArchive(path)
		if err != nil {
			fileio.WriteErrorSidecar(path, err)
			_ = beeep.Alert("CsvWatcher", "Archivieren fehlgeschlagen: "+err.Error(), "")
			return
		}
		w.MarkSelfWritten(path)
		if err := fileio.WriteCSV(path, rows); err != nil {
			_ = os.Rename(archivePath, path) // rollback
			fileio.WriteErrorSidecar(path, err)
			_ = beeep.Alert("CsvWatcher", "Schreiben fehlgeschlagen: "+err.Error(), "")
			return
		}
		_ = beeep.Notify("CsvWatcher", filepath.Base(path)+" umgewandelt. Original in verarbeitet/.", "")
		return
	}

	// manual convert outside watched folder → neue Datei mit Suffix
	ext := filepath.Ext(path)
	outPath := path[:len(path)-len(ext)] + "_konvertiert" + ext
	if err := fileio.WriteCSV(outPath, rows); err != nil {
		_ = beeep.Alert("CsvWatcher", "Schreiben fehlgeschlagen: "+err.Error(), "")
		return
	}
	_ = beeep.Notify("CsvWatcher", "Neue Datei: "+filepath.Base(outPath), "")
}

func openInFileExplorer(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("explorer", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	_ = cmd.Start()
}
