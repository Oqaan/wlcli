package main

import (
	"slices"
	"testing"
)

var testRows = [][]string{
	{"StopID", "DIVA", "StopText"},
	{"4181", "", "Kagraner Platz"},
	{"885", "", "Kagraner Platz U"},
	{"4116", "", "Schwedenplatz"},
	{"4127", "", "Kagran"},
}

func TestFindStopIDs(t *testing.T) {
	tests := []struct {
		name   string
		search string
		want   []string
	}{
		{"exact name", "Schwedenplatz", []string{"4116"}},
		{"partial name", "Kagraner Platz", []string{"4181", "885"}},
		{"lowercase", "schwedenplatz", []string{"4116"}},
		{"uppercase", "SCHWEDENPLATZ", []string{"4116"}},
		{"mixed case", "SchWeDeNpLaTz", []string{"4116"}},
		{"no match", "Nonexistent Stop", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findStopIDs(testRows, tt.search)
			if !slices.Equal(got, tt.want) {
				t.Errorf("findStopIDs(%q) = %v, want %v", tt.search, got, tt.want)
			}
		})
	}
}
