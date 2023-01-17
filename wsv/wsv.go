package wsv

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

type WsvMeasurement struct {
	Value       float64
	Timestamp   time.Time
	StateMnwMhw string `json:"stateMnwMhw,omitempty"`
	StateNswHsw string `json:"stateNswHsw,omitempty"`
}

type WsvSeries struct {
	Shortname          string
	Unit               string
	Equidistance       int
	CurrentMeasurement WsvMeasurement
}

type WsvLevelData struct {
	Number     string
	Kilometer  float64 `json:"km"`
	TimeSeries [3]WsvSeries
}

func decodeJSON(data []byte) (*WsvLevelData, error) {
	var result WsvLevelData
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&result); err != nil {
		log.Printf("Error decoding JSON data '%s': %v", data, err)
		return nil, err
	}
	return &result, nil
}

func fetchWsvJSON() ([]byte, error) {
	// fetch pegel from https://www.pegelonline.wsv.de/
	uuid := "a6ee8177-107b-47dd-bcfd-30960ccc6e9c"
	url := "https://www.pegelonline.wsv.de/webservices/rest-api/v2/stations/" + uuid + ".json?includeTimeseries=true&includeCurrentMeasurement=true"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error getting url %v: %v\n", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v}\n", err)
		return nil, err
	}
	return data, nil
}

func QueryPegelOnline() (*WsvLevelData, error) {

	data, err := fetchWsvJSON()
	if err != nil {
		return nil, err
	}

	wsv, err := decodeJSON(data)
	if err != nil {
		return nil, err
	}

	return wsv, nil
}
