// Package fileio handles CP1252-encoded DATEV CSVs: read, write, and archival.
package fileio

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// ReadCSV reads a CP1252 (Windows-1252) encoded semicolon-separated CSV.
func ReadCSV(path string) ([][]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	utf8Bytes, _, err := transform.Bytes(charmap.Windows1252.NewDecoder(), raw)
	if err != nil {
		return nil, fmt.Errorf("decode cp1252 %s: %w", path, err)
	}
	r := csv.NewReader(bytes.NewReader(utf8Bytes))
	r.Comma = ';'
	r.LazyQuotes = true
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse csv %s: %w", path, err)
	}
	return rows, nil
}

// WriteCSV writes rows back to path as CP1252, semicolon-separated, CRLF line endings.
func WriteCSV(path string, rows [][]string) error {
	var utf8Buf bytes.Buffer
	w := csv.NewWriter(&utf8Buf)
	w.Comma = ';'
	w.UseCRLF = true
	if err := w.WriteAll(rows); err != nil {
		return fmt.Errorf("encode csv: %w", err)
	}
	cp1252Bytes, _, err := transform.Bytes(charmap.Windows1252.NewEncoder(), utf8Buf.Bytes())
	if err != nil {
		return fmt.Errorf("encode cp1252: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, cp1252Bytes, 0o644); err != nil {
		return fmt.Errorf("write tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename tmp to %s: %w", path, err)
	}
	return nil
}

// MoveToArchive moves the source file into an archive subdirectory named "verarbeitet"
// sitting next to the source. If the destination filename already exists, a timestamp
// suffix is appended. Returns the final archive path.
func MoveToArchive(sourcePath string) (string, error) {
	dir := filepath.Dir(sourcePath)
	base := filepath.Base(sourcePath)
	archiveDir := filepath.Join(dir, "verarbeitet")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", archiveDir, err)
	}
	dest := filepath.Join(archiveDir, base)
	if _, err := os.Stat(dest); err == nil {
		ext := filepath.Ext(base)
		name := base[:len(base)-len(ext)]
		stamp := time.Now().Format("20060102_150405")
		dest = filepath.Join(archiveDir, fmt.Sprintf("%s_%s%s", name, stamp, ext))
	}
	if err := os.Rename(sourcePath, dest); err != nil {
		// Fallback for cross-device moves (rare, but safe).
		if copyErr := copyFile(sourcePath, dest); copyErr != nil {
			return "", fmt.Errorf("move %s -> %s: %w", sourcePath, dest, err)
		}
		if rmErr := os.Remove(sourcePath); rmErr != nil {
			return "", fmt.Errorf("remove source %s: %w", sourcePath, rmErr)
		}
	}
	return dest, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// WriteErrorSidecar drops a .error.log next to the source explaining why conversion failed.
func WriteErrorSidecar(sourcePath string, cause error) {
	logPath := sourcePath + ".error.log"
	msg := fmt.Sprintf("[%s] %v\n", time.Now().Format(time.RFC3339), cause)
	_ = os.WriteFile(logPath, []byte(msg), 0o644)
}
