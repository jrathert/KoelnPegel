package wsv

import (
	"strings"
	"testing"
)

func TestFetchWsvJSON(t *testing.T) {
	data, err := fetchWsvJSON()
	if err != nil {
		t.Errorf(`fetchWsvJSON() returns error (%v)`, err)
	}
	if !strings.Contains(string(data), "a6ee8177-107b-47dd-bcfd-30960ccc6e9c") {
		t.Errorf(`fetchWsvJSON() does not contain correct uuid (%v)`, err)
	}
}

func TestDecodeJSON(t *testing.T) {
	wsv, err := decodeJSON([]byte(jsondata))
	if err != nil {
		t.Errorf(`decodeJSON(jsondata) returns error (%v)`, err)
	}
	if wsv.Kilometer != 688 || len(wsv.TimeSeries) != 3 || wsv.TimeSeries[0].CurrentMeasurement.Value != 475 {
		t.Errorf(`decodeJSON(jsondata) does not contain proper data (%v)`, wsv)
	}

	wsv, err = decodeJSON([]byte("{}"))
	if err != nil {
		t.Errorf(`decodeJSON("") returns error (%v)`, err)
	}
	if wsv.Kilometer != 0 {
		t.Errorf(`decodeJSON("") wsvKilometer != 0 (%v)`, wsv.Kilometer)
	}
}

// func TestNonDecodeJSON(t *testing.T) {
// 	wsv, err := decodeJSON([]byte("{]"))
// 	if err == nil {
// 		t.Errorf(`decodeJSON("{]") does not return error`)
// 	}
// 	if wsv != nil {
// 		t.Errorf(`decodeJSON("{]") does not return nil data`)
// 	}
// }

const jsondata = `
	{
		"uuid": "a6ee8177-107b-47dd-bcfd-30960ccc6e9c",
		"number": "2730010",
		"shortname": "KÖLN",
		"longname": "KÖLN",
		"km": 688.0,
		"agency": "STANDORT KÖLN",
		"longitude": 6.963300159749651,
		"latitude": 50.93694929574385,
		"water": {
		  "shortname": "RHEIN",
		  "longname": "RHEIN"
		},
		"timeseries": [
		  {
			"shortname": "W",
			"longname": "WASSERSTAND ROHDATEN",
			"unit": "cm",
			"equidistance": 15,
			"currentMeasurement": {
			  "timestamp": "2023-01-12T16:45:00+01:00",
			  "value": 475.0,
			  "stateMnwMhw": "normal",
			  "stateNswHsw": "normal"
			},
			"gaugeZero": {
			  "unit": "m. ü. NHN",
			  "value": 35.038,
			  "validFrom": "2019-11-01"
			}
		  },
		  {
			"shortname": "Q",
			"longname": "ABFLUSS_ROHDATEN",
			"unit": "m³/s",
			"equidistance": 15,
			"currentMeasurement": {
			  "timestamp": "2023-01-12T16:30:00+01:00",
			  "value": 3230.0
			}
		  },
		  {
			"shortname": "WT",
			"longname": "WASSERTEMPERATUR ROHDATEN",
			"unit": "°C",
			"equidistance": 15,
			"currentMeasurement": {
			  "timestamp": "2023-01-12T16:45:00+01:00",
			  "value": 7.0
			}
		  }
		]
	  }

`
