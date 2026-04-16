package fileio_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"csvwatcher/internal/fileio"
	"csvwatcher/internal/transform"

	"golang.org/x/text/encoding/charmap"
)

func TestEndToEnd_RealExampleFile(t *testing.T) {
	src := filepath.Join("..", "..", "original_EXTF_Div-Adressen_DEMOOPH_all_2026-03-25_05-13.csv")
	if _, err := os.Stat(src); err != nil {
		t.Skipf("fixture missing: %v", err)
	}

	rows, err := fileio.ReadCSV(src)
	if err != nil {
		t.Fatalf("ReadCSV: %v", err)
	}
	if len(rows) < 3 {
		t.Fatalf("expected ≥3 rows, got %d", len(rows))
	}
	if rows[1][0] != "Adressnummer" {
		t.Fatalf("row 2 col A = %q, want %q", rows[1][0], "Adressnummer")
	}
	if rows[2][0] != "500192" {
		t.Fatalf("row 3 col A = %q, want %q", rows[2][0], "500192")
	}
	if rows[2][1] != "Traumschloss AG" {
		t.Fatalf("row 3 col B = %q, want %q", rows[2][1], "Traumschloss AG")
	}

	colCount := len(rows[1])
	row1Copy := append([]string(nil), rows[0]...)

	transform.Apply(rows)

	if len(rows[0]) != len(row1Copy) {
		t.Fatalf("row 1 length changed: %d vs %d", len(rows[0]), len(row1Copy))
	}
	for i, v := range row1Copy {
		if rows[0][i] != v {
			t.Fatalf("row 1 col %d: got %q, want %q", i, rows[0][i], v)
		}
	}

	if rows[1][0] != "" {
		t.Errorf("row 2 A not cleared: %q", rows[1][0])
	}
	if rows[1][1] != "Konto" {
		t.Errorf("row 2 B = %q, want unchanged %q", rows[1][1], "Konto")
	}
	if len(rows[1]) != colCount {
		t.Errorf("row 2 column count changed: %d vs %d", len(rows[1]), colCount)
	}

	if rows[2][0] != "" {
		t.Errorf("row 3 A not cleared: %q", rows[2][0])
	}
	if rows[2][1] != "500192" {
		t.Errorf("row 3 B = %q, want %q", rows[2][1], "500192")
	}
	if rows[2][3] != "Traumschloss AG" {
		t.Errorf("row 3 D = %q, want %q", rows[2][3], "Traumschloss AG")
	}

	tmp := t.TempDir()
	out := filepath.Join(tmp, "out.csv")
	if err := fileio.WriteCSV(out, rows); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}

	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	// Verify CP1252: decoding the bytes should yield valid UTF-8 with umlauts.
	decoded, err := charmap.Windows1252.NewDecoder().Bytes(raw)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !bytes.Contains(decoded, []byte("Adressattyp Unternehmen")) {
		t.Error("expected preserved umlauts in header (Adressattyp Unternehmen)")
	}
	if !bytes.Contains(raw, []byte{0x0d, 0x0a}) {
		t.Error("expected CRLF line endings in output")
	}
}

func TestMoveToArchive_CreatesSubfolderAndMoves(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "test.csv")
	if err := os.WriteFile(src, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest, err := fileio.MoveToArchive(src)
	if err != nil {
		t.Fatalf("MoveToArchive: %v", err)
	}
	if filepath.Dir(dest) != filepath.Join(tmp, "verarbeitet") {
		t.Errorf("dest dir = %q, want verarbeitet/", filepath.Dir(dest))
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("source still exists after move")
	}
	if _, err := os.Stat(dest); err != nil {
		t.Errorf("dest missing: %v", err)
	}
}

func TestMoveToArchive_CollisionAddsTimestamp(t *testing.T) {
	tmp := t.TempDir()
	archive := filepath.Join(tmp, "verarbeitet")
	_ = os.MkdirAll(archive, 0o755)
	_ = os.WriteFile(filepath.Join(archive, "test.csv"), []byte("existing"), 0o644)

	src := filepath.Join(tmp, "test.csv")
	_ = os.WriteFile(src, []byte("new"), 0o644)

	dest, err := fileio.MoveToArchive(src)
	if err != nil {
		t.Fatalf("MoveToArchive: %v", err)
	}
	if filepath.Base(dest) == "test.csv" {
		t.Error("expected timestamped filename, got plain test.csv")
	}
}
