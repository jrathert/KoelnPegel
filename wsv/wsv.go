package wsv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func FetchPegelOnline() (*WsvLevelData, error) {

	// fetch pegel from https://www.pegelonline.wsv.de/
	uuid := "a6ee8177-107b-47dd-bcfd-30960ccc6e9c"
	url := "https://www.pegelonline.wsv.de/webservices/rest-api/v2/stations/" + uuid + ".json?includeTimeseries=true&includeCurrentMeasurement=true"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching data: ", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading body: ", err)
		return nil, err
	}
	// fmt.Println(string(body))

	var result WsvLevelData
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
		fmt.Println("Error decoding json data:", string(body))
		return nil, err
	}
	return &result, nil
}

