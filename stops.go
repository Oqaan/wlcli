package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
)

// maxStopIDs limits how many stop IDs are sent in one request, because the API rejects requests with too many
const maxStopIDs = 100

// loadStops downloads the stops CSV and returns all rows
func loadStops() ([][]string, error) {
	resp, err := client.Get("https://www.wienerlinien.at/ogd_realtime/doku/ogd/wienerlinien-ogd-haltepunkte.csv")
	if err != nil {
		return nil, fmt.Errorf("downloading stops: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading stops: unexpected status %s", resp.Status)
	}
	reader := csv.NewReader(resp.Body)
	reader.Comma = ';' // The Wiener Linien CSV uses semicolons instead of commas
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading stops CSV: %w", err)
	}
	return rows, nil
}

// findStopIDs returns the stop IDs of all stops matching name (case-insensitive)
// Stops named exactly name or "name U" are preferred. If there are none, 
// all stops containing name are returned
func findStopIDs(rows [][]string, name string) []string {
	var exact []string
	var partial []string
	search := strings.ToLower(name)

	for _, record := range rows[1:] { // Skip the header row (rows[0])
		// CSV columns: 0 = StopID, 2 = StopText (name)
		stopName := strings.ToLower(record[2])

		// U-Bahn platforms are named "X", tram and bus stops at the same station "X U"
		if stopName == search || stopName == search+" u" {
			exact = append(exact, record[0])
		} else if strings.Contains(stopName, search) {
			partial = append(partial, record[0])
		}
	}
	if len(exact) > 0 {
		return exact
	}
	return partial
}
