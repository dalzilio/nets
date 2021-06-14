// Copyright (c) 2021 Silvano DAL ZILIO
//
// GNU Affero GPL v3

package nets

import (
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	tables := []struct {
		file   string
		pl, tr int
	}{
		{"abp.net", 12, 16},
		{"demo.net", 4, 7},
		{"ifip.net", 5, 5},
		{"sokoban_3.net", 410, 452},
	}
	for _, v := range tables {
		file, err := os.Open("testdata/" + v.file)
		if err != nil {
			t.Errorf("Error opening file %s; %s", v.file, err)
		}
		expected, err := Parse(file)
		if err != nil {
			t.Errorf("Error parsing file %s; %s", v.file, err)
		}
		if pl := len(expected.Pl); pl != v.pl {
			t.Errorf("Wrong number of places in %s, expected %d, actual %d", v.file, v.pl, pl)
		}
		if tr := len(expected.Tr); tr != v.tr {
			t.Errorf("Wrong number of transitions in %s, expected %d, actual %d", v.file, v.tr, tr)
		}
	}
}
