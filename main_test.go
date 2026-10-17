package main

import (
	"image/color"
	"slices"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestLineColor(t *testing.T) {
	tests := []struct {
		name     string
		lineName string
		want     color.Color
	}{
		{"U1 red", "U1", lipgloss.Color("#E3000F")},
		{"U4 green", "U4", lipgloss.Color("#00963F")},
		{"tram has no color", "26", lipgloss.NoColor{}},
		{"bus has no color", "13A", lipgloss.NoColor{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lineColor(tt.lineName)
			if got != tt.want {
				t.Errorf("lineColor(%q) = %v, want %v", tt.lineName, got, tt.want)
			}
		})
	}
}

func TestCompareLines(t *testing.T) {
	lines := []string{"N25", "13A", "U4", "D", "4A", "2", "U1", "13B"}
	want := []string{"U1", "U4", "D", "2", "4A", "13A", "13B", "N25"}

	slices.SortFunc(lines, compareLines)

	if !slices.Equal(lines, want) {
		t.Errorf("sorted lines = %v, want %v", lines, want)
	}
}

func TestRelevantTrafficInfos(t *testing.T) {
	infos := []TrafficInfo{
		{Title: "elevator here", RelatedLines: []string{"U1"}, RelatedStops: []int{4116}},
		{Title: "elevator elsewhere", RelatedLines: []string{"U1"}, RelatedStops: []int{9999}},
		{Title: "works on shown line", RelatedLines: []string{"2"}},
		{Title: "works on other line", RelatedLines: []string{"D"}},
	}
	ids := []string{"4116"}
	lines := []string{"U1", "2"}

	got := relevantTrafficInfos(infos, ids, lines)

	var gotTitles []string
	for _, info := range got {
		gotTitles = append(gotTitles, info.Title)
	}
	want := []string{"elevator here", "works on shown line"}
	if !slices.Equal(gotTitles, want) {
		t.Errorf("relevant titles = %v, want %v", gotTitles, want)
	}
}
