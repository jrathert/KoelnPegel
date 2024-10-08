package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"memcpy.me/kpg/wsv"
)

const KATA_02 = 1130 // 1130 - Altstadt läuft voll
const KATA_01 = 1000 // 1000 - Rheinufertunnel gesperrt
const MARK_02 = 830  //  830 - Warnstufe 2
const MARK_01 = 620  //  620 - Warnstufe 1
const GLW = 139      //  139 - GLW: Gleichwertiger Wasserstand

const STARK = 10 // 5 - starke Änderung cm / Zeit
const MID = 5    // 3 - Änderung cm / Zeit
const LOW = 5    // 1 - leichte Änderung cm / Zeit

const LAST_FILE = ".kpg_last"

// Measurement represents in compact form one measurement of the Cologne level
// It has a timestamp, a water level and a water temperature
type Measurement struct {
	Timestamp   time.Time
	Level       float64
	Temperature float64
}

// String representation of a measurement
func (m Measurement) String() string {
	return fmt.Sprintf("%v: %v cm (%v °C)", m.Timestamp, m.Level, m.Temperature)
}

func retrieveCurrentData() Measurement {
	leveldata, err := wsv.QueryPegelOnline()
	if err != nil {
		log.Println("No current data available")
		return Measurement{}
	}
	return Measurement{
		Timestamp:   leveldata.TimeSeries[0].CurrentMeasurement.Timestamp,
		Level:       leveldata.TimeSeries[0].CurrentMeasurement.Value,
		Temperature: leveldata.TimeSeries[2].CurrentMeasurement.Value,
	}
}

func getTendencyString(diff float64) (string, rune) {
	// zur Tendenz siehe: https://undine.bafg.de/rhein/zustand-aktuell/rhein_akt_WQ.html
	//   Diffzeitraum: 4h
	//   stark: mehr als +/-10 cm Veränderung in 4h
	//   normal_ mehr als +/-5 cm Veränderung in 4h
	//   konstant: bis zu +/- 5cm Veränderung in 4h

	if diff > STARK {
		return "stark steigend", '\u2197' // ↗
	} else if diff > MID {
		return "steigend", '\u2197' // ↗
	} else if diff > LOW {
		return "leicht steigend", '\u2197' // ↗
	} else if diff < -STARK {
		return "stark fallend", '\u2198' // ↘
	} else if diff < -MID {
		return "fallend", '\u2198' //  ↘
	} else if diff < -LOW {
		return "leicht fallend", '\u2198' // ↘
	} else {
		return "konstant", '\u27a1' // ➡
	}
}

func prepareStatusString(current Measurement, trend string, icon rune) string {

	var sb strings.Builder

	wtime := current.Timestamp.Format("15:04")

	// sb.WriteString(fmt.Sprintf("Stand am Pegel Köln um %v Uhr (%v °C): %v cm", wtime, current.Temperature, current.Level))
	sb.WriteString(fmt.Sprintf("Stand am Pegel Köln um %v Uhr: %v cm", wtime, current.Level))
	if icon != 0 {
		sb.WriteString(fmt.Sprintf(" - %v %v", trend, string(icon)))
	}
	sb.WriteString("\n")
	if current.Level >= KATA_02 {
		sb.WriteString("\u26a0 Überflutung der Altstadt \U0001f30a, Rheinufertunnel gesperrt \U0001f6a7\n")
	} else if current.Level >= KATA_01 {
		sb.WriteString("\u26a0 Rheinufertunnel gesperrt \U0001f6a7\n")
	}
	if current.Level >= MARK_02 {
		sb.WriteString("\u26a0 Hochwassermarke 2 erreicht - Schiffsverkehr gesperrt \U0001f6e5\n")
	} else if current.Level >= MARK_01 {
		sb.WriteString("\u26a0 Hochwassermarke 1 erreicht - Schiffsverkehr verlangsamt \U0001f6e5\n")
	}
	if current.Level < GLW {
		sb.WriteString("\u26a0 Unter gleichwertigem Wasserstand - Schiffsverkehr ggf. eingeschränkt\n")
	}
	sb.WriteString("Mehr Infos: \U0001f449 https://www.koeln.de/wetter/rheinpegel/")

	return sb.String()
}

func savePostTime(measurement Measurement) {
	val, err := measurement.Timestamp.MarshalText()
	if err != nil {
		log.Println("Error encoding timestamp:", err)
		return
	}
	err = os.WriteFile(LAST_FILE, val, 0664)
	if err != nil {
		log.Println("Error saving file", err)
	}
}

func lastPostTime() (time.Time, error) {
	ts := time.Unix(0, 0)
	data, err := os.ReadFile(LAST_FILE)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Println("Error reading file:", err)
			return ts, err
		} else {
			return ts, nil
		}
	}
	err = ts.UnmarshalText(data)
	if err != nil {
		log.Println("Error decoding timestamp:", err)
		return ts, err
	}
	return ts, nil
}

func checkIfPostNow(measurement Measurement) bool {

	lastPost, err := lastPostTime()
	if err != nil {
		return false
	}
	diffMin := int(measurement.Timestamp.Sub(lastPost).Minutes())
	if measurement.Level > KATA_02 {
		// post all available data!
		return true
	} else if measurement.Level > KATA_01 {
		// all 30 minutes
		return diffMin >= 30 && measurement.Timestamp.Minute()%30 == 0 // 30 -> 30
	} else if measurement.Level > MARK_02 {
		// all 60 minutes
		return diffMin >= 60 && measurement.Timestamp.Minute() == 0 // 60 -> 0 (or 60!)
	} else if measurement.Level > MARK_01 {
		// all 120 minutes
		return diffMin >= 120 && measurement.Timestamp.Minute() == 0 && measurement.Timestamp.Hour()%2 == 0 // 12
	} else {
		// all 4 hours
		return diffMin >= 240 && measurement.Timestamp.Minute() == 0 && measurement.Timestamp.Hour()%4 == 0
	}
}
