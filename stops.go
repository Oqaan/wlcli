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

// findStopIDs returns the stop IDs of all rows whose name contains name (case-insensitive)
func findStopIDs(rows [][]string, name string) []string {
	var ids []string
	search := strings.ToLower(name)
	for _, record := range rows {
		// CSV columns: 0 = StopID, 2 = StopText (name)
		if strings.Contains(strings.ToLower(record[2]), search) {
			ids = append(ids, record[0])
		}
	}
	return ids
}
