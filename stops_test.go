package main

import (
	"slices"
	"testing"
)

// testRows is a small fake version of the stops CSV
// Only column 0 (StopID) and column 2 (StopText) are used by findStopIDs
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
		{"header is not a stop", "text", []string{}},
		{"exact match preferred", "kagran", []string{"4127"}},
		{"partial match as fallback", "kagraner", []string{"4181", "885"}},
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
