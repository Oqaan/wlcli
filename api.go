package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type MonitorResponse struct {
	Data    Data    `json:"data"`
	Message Message `json:"message"`
}

type Message struct {
	Value string `json:"value"`
}

type Data struct {
	Monitors []Monitor `json:"monitors"`
}

type Monitor struct {
	Lines []Line `json:"lines"`
}

type Line struct {
	Name       string     `json:"name"`
	Towards    string     `json:"towards"`
	Departures Departures `json:"departures"`
}

// Departures is an extra wrapper object in the API response around the departure list
type Departures struct {
	Departure []Departure `json:"departure"`
}

type Departure struct {
	DepartureTime DepartureTime `json:"departureTime"`
}

type DepartureTime struct {
	Countdown int `json:"countdown"`
}

// client is a shared HTTP client with a timeout to avoid hanging requests
var client = &http.Client{Timeout: 10 * time.Second}

// fetchDepartures requests the real-time departures for the given stop IDs
func fetchDepartures(ids []string) (MonitorResponse, error) {
	stopIDs := strings.Join(ids, "&stopId=")
	apiURL := "https://www.wienerlinien.at/ogd_realtime/monitor?stopId=" + stopIDs

	resp, err := client.Get(apiURL)
	if err != nil {
		return MonitorResponse{}, fmt.Errorf("fetching departures: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return MonitorResponse{}, fmt.Errorf("reading departure response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		var apiResult MonitorResponse
		_ = json.Unmarshal(body, &apiResult) // Try to get the API's error message; the body may not be JSON (e.g. an HTML error page), so ignore errors.
		if apiResult.Message.Value != "" {
			return MonitorResponse{}, fmt.Errorf("fetching departures: %s: %s", resp.Status, apiResult.Message.Value)
		}
		return MonitorResponse{}, fmt.Errorf("fetching departures: unexpected status %s", resp.Status)
	}
	var result MonitorResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return MonitorResponse{}, fmt.Errorf("unmarshaling departure response: %w", err)
	}
	return result, nil
}
