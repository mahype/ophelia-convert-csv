package transform

import (
	"reflect"
	"testing"
)

func TestApply_Row1Unchanged(t *testing.T) {
	rows := [][]string{
		{"EXTF", "700", "16", "Debitoren/Kreditoren"},
		{"Adressnummer", "Konto", "Anrede", "Name"},
		{"500192", "Traumschloss AG", "", ""},
	}
	original := append([]string(nil), rows[0]...)
	Apply(rows)
	if !reflect.DeepEqual(rows[0], original) {
		t.Fatalf("row 1 changed: got %v, want %v", rows[0], original)
	}
}

func TestApply_Row2ReplacedByTargetHeader(t *testing.T) {
	rows := [][]string{
		{"EXTF"},
		{"Adressnummer", "Konto", "Anrede", "Name"},
		{"500192", "Traumschloss AG", "", ""},
	}
	Apply(rows)
	if len(rows[1]) != TargetCols {
		t.Fatalf("row 2 length = %d, want %d", len(rows[1]), TargetCols)
	}
	if !reflect.DeepEqual(rows[1], TargetHeader) {
		t.Fatalf("row 2 is not TargetHeader")
	}
}

func TestApply_Row2LandmarkCells(t *testing.T) {
	rows := [][]string{
		{"EXTF"},
		{"Adressnummer", "Konto"},
	}
	Apply(rows)
	checks := map[int]string{
		0:   "Konto",
		1:   "Name (Adressattyp Unternehmen)",
		8:   "EU-Land",
		9:   "EU-UStID",
		95:  "Leerfeld 1",
		203: "SWIFTCode 9",
		221: "SEPA Mandatsreferenz 1",
		253: "Letzte Frist",
	}
	for i, want := range checks {
		if rows[1][i] != want {
			t.Errorf("row 2 col %d = %q, want %q", i+1, rows[1][i], want)
		}
	}
}

func TestApply_DataRowUnchangedAndPadded(t *testing.T) {
	rows := [][]string{
		{"EXTF"},
		{"Adressnummer", "Konto"},
		{"500192", "Traumschloss AG", "", "", "Rest5", "Rest6"},
	}
	Apply(rows)
	if len(rows[2]) != TargetCols {
		t.Fatalf("data row length = %d, want %d", len(rows[2]), TargetCols)
	}
	wantHead := []string{"500192", "Traumschloss AG", "", "", "Rest5", "Rest6"}
	if !reflect.DeepEqual(rows[2][:6], wantHead) {
		t.Fatalf("row 3 head: got %v, want %v", rows[2][:6], wantHead)
	}
	for i := 6; i < TargetCols; i++ {
		if rows[2][i] != "" {
			t.Errorf("row 3 col %d = %q, want empty", i+1, rows[2][i])
		}
	}
}

func TestApply_ShortRowIsPaddedTo254(t *testing.T) {
	rows := [][]string{
		{"EXTF"},
		{"Adressnummer", "Konto"},
		{"500192", "Traumschloss AG"},
	}
	Apply(rows)
	if len(rows[2]) != TargetCols {
		t.Fatalf("short row not padded: len = %d, want %d", len(rows[2]), TargetCols)
	}
	if rows[2][0] != "500192" || rows[2][1] != "Traumschloss AG" {
		t.Errorf("row 3 head changed unexpectedly: %v", rows[2][:2])
	}
}

func TestApply_MultipleDataRows(t *testing.T) {
	rows := [][]string{
		{"EXTF"},
		{"Adressnummer", "Konto"},
		{"500192", "Firma A", "", "", "X"},
		{"500193", "Firma B", "", "", "Y"},
		{"500194", "Firma C", "", "", "Z"},
	}
	Apply(rows)

	for i, want := range []struct {
		a, b, e string
	}{
		{"500192", "Firma A", "X"},
		{"500193", "Firma B", "Y"},
		{"500194", "Firma C", "Z"},
	} {
		row := rows[i+2]
		if len(row) != TargetCols {
			t.Errorf("row %d length = %d, want %d", i+2, len(row), TargetCols)
		}
		if row[0] != want.a || row[1] != want.b || row[4] != want.e {
			t.Errorf("row %d head wrong: %v", i+2, row[:5])
		}
	}
}

func TestApply_EmptyInput(t *testing.T) {
	rows := [][]string{}
	got := Apply(rows)
	if len(got) != 0 {
		t.Fatalf("empty input should yield empty output, got %v", got)
	}
}

func TestTargetHeader_HasExpectedLength(t *testing.T) {
	if len(TargetHeader) != TargetCols {
		t.Fatalf("TargetHeader len = %d, want %d", len(TargetHeader), TargetCols)
	}
}
