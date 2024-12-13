package nets

import (
	"fmt"
	"os"
	"testing"
)

func TestMarkingAddToPlace(t *testing.T) {
	tables := []struct {
		Marking
		pl       int
		mult     int
		expected Marking
	}{
		{Marking{}, 2, 6, Marking{Atom{2, 6}}},
		{Marking{Atom{3, 4}}, 3, 6, Marking{Atom{3, 10}}},
		{Marking{Atom{4, 4}}, 3, 0, Marking{Atom{4, 4}}},
		{Marking{Atom{4, 4}}, 4, -4, Marking{}},
		{Marking{Atom{4, 4}}, 3, 2, Marking{Atom{3, 2}, Atom{4, 4}}},
		{Marking{Atom{0, -1}, Atom{5, 4}}, 5, -1, Marking{Atom{0, -1}, Atom{5, 3}}},
		{Marking{Atom{6, 7}, Atom{8, 7}, Atom{10, 4}}, 8, -7, Marking{Atom{6, 7}, Atom{10, 4}}},
	}

	for _, tt := range tables {
		actual := tt.Marking.AddToPlace(tt.pl, tt.mult)
		fmt.Println(tt.expected)
		if !actual.Equal(tt.expected) {
			t.Errorf("%v .AddToPlace(%d, %d): expected %v, actual %v", tt.Marking, tt.pl, tt.mult, tt.expected, actual)
		}
	}
}

func TestMtoa(t *testing.T) {
	file, err := os.Open("testdata/ifip.net")
	if err != nil {
		t.Errorf("Error opening file testdata/ifip.net; %s", err)
	}
	net, err := Parse(file)
	if err != nil {
		t.Errorf("Error parsing file testdata/ifip.net; %s", err)
	}

	tables := []struct {
		Marking
		expected string
	}{
		{Marking{}, ""},
		{Marking{Atom{3, 4}}, "p4*4"},
		{Marking{Atom{0, -1}, Atom{4, 4}}, "p1*-1 p5*4"},
		{Marking{Atom{0, 7}, Atom{2, 1}, Atom{3, 4}}, "p1*7 p3 p4*4"},
	}

	for _, tt := range tables {
		actual := net.Mtoa(tt.Marking)
		if actual != tt.expected {
			t.Errorf("net.Mtoa(%d): expected %v, actual %v", tt.Marking, tt.expected, actual)
		}
	}
}
