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
}

func TestFindStopIDs(t *testing.T) {
	got := findStopIDs(testRows, "Schwedenplatz")

	want := []string{"4116"}
	if !slices.Equal(got, want) {
		t.Errorf("findStopIDs(%q) = %v, want %v", "Schwedenplatz", got, want)
	}
}

func TestFindStopIDsPartialName(t *testing.T) {
	got := findStopIDs(testRows, "Kagraner Platz")
	want := []string{"4181", "885"}
	if !slices.Equal(got, want) {
		t.Errorf("findStopIDs(%q) = %v, want %v", "Kagraner Platz", got, want)
	}
}

func TestFindStopIDsLowercase(t *testing.T) {
	got := findStopIDs(testRows, "schwedenplatz")
	want := []string{"4116"}
	if !slices.Equal(got, want) {
		t.Errorf("findStopIDs(%q) = %v, want %v", "schwedenplatz", got, want)
	}
}

func TestFindStopIDsNoMatch(t *testing.T) {
	got := findStopIDs(testRows, "Nonexistent Stop")
	want := []string{}
	if !slices.Equal(got, want) {
		t.Errorf("findStopIDs(%q) = %v, want %v", "Nonexistent Stop", got, want)
	}
}
