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

func TestApply_Row2ClearsA(t *testing.T) {
	rows := [][]string{
		{"EXTF"},
		{"Adressnummer", "Konto", "Anrede", "Name"},
		{"500192", "Traumschloss AG", "", ""},
	}
	Apply(rows)
	want := []string{"", "Konto", "Anrede", "Name"}
	if !reflect.DeepEqual(rows[1], want) {
		t.Fatalf("row 2: got %v, want %v", rows[1], want)
	}
}

func TestApply_DataRowShift(t *testing.T) {
	rows := [][]string{
		{"EXTF"},
		{"Adressnummer", "Konto", "Anrede", "Name"},
		{"500192", "Traumschloss AG", "Herr", "Alter-Name", "Rest5", "Rest6"},
	}
	Apply(rows)
	want := []string{"", "500192", "Herr", "Traumschloss AG", "Rest5", "Rest6"}
	if !reflect.DeepEqual(rows[2], want) {
		t.Fatalf("row 3: got %v, want %v", rows[2], want)
	}
}

func TestApply_ShortRowIsPadded(t *testing.T) {
	rows := [][]string{
		{"EXTF"},
		{"Adressnummer", "Konto"},
		{"500192", "Traumschloss AG"},
	}
	Apply(rows)
	want := []string{"", "500192", "", "Traumschloss AG"}
	if !reflect.DeepEqual(rows[2], want) {
		t.Fatalf("short row: got %v, want %v", rows[2], want)
	}
}

func TestApply_MultipleDataRows(t *testing.T) {
	rows := [][]string{
		{"EXTF"},
		{"Adressnummer", "Konto", "Anrede", "Name"},
		{"500192", "Firma A", "", "", "X"},
		{"500193", "Firma B", "", "", "Y"},
		{"500194", "Firma C", "", "", "Z"},
	}
	Apply(rows)

	for i, wantRow := range [][]string{
		{"", "500192", "", "Firma A", "X"},
		{"", "500193", "", "Firma B", "Y"},
		{"", "500194", "", "Firma C", "Z"},
	} {
		if !reflect.DeepEqual(rows[i+2], wantRow) {
			t.Errorf("row %d: got %v, want %v", i+2, rows[i+2], wantRow)
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
