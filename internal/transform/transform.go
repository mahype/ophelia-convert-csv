// Package transform implements the DATEV column shift specified by the user:
//
//	Zeile 1:   unverändert (EXTF-Metadaten)
//	Zeile 2:   Spalte A geleert, Rest unverändert
//	Zeile 3+:  D := B; B := A; A := ""  (A=0, B=1, C=2, D=3)
package transform

func Apply(rows [][]string) [][]string {
	for i := range rows {
		if i == 0 {
			continue
		}
		if i == 1 {
			if len(rows[i]) > 0 {
				rows[i][0] = ""
			}
			continue
		}
		for len(rows[i]) < 4 {
			rows[i] = append(rows[i], "")
		}
		rows[i][3] = rows[i][1]
		rows[i][1] = rows[i][0]
		rows[i][0] = ""
	}
	return rows
}
