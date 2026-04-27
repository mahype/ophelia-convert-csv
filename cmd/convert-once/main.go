package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"csvwatcher/internal/fileio"
	"csvwatcher/internal/transform"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: convert-once <input.csv> [output.csv]")
		os.Exit(2)
	}
	in := os.Args[1]

	var out string
	if len(os.Args) >= 3 {
		out = os.Args[2]
	} else {
		ext := filepath.Ext(in)
		base := strings.TrimSuffix(in, ext)
		out = base + "_konvertiert" + ext
	}

	rows, err := fileio.ReadCSV(in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read:", err)
		os.Exit(1)
	}
	transform.Apply(rows)
	if err := fileio.WriteCSV(out, rows); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		os.Exit(1)
	}
	fmt.Printf("OK  %s  ->  %s  (%d Zeilen)\n", in, out, len(rows))
}
