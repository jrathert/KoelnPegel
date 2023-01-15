package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const KATA_02 = 1130 // 1130 - Altstadt läuft voll
const KATA_01 = 1000 // 1000 - Rheinufertunnel gesperrt
const MARK_02 = 830  //  830 - Warnstufe 2
const MARK_01 = 620  //  620 - Warnstufe 1
const GLW = 139      //  139 - GLW: Gleichwertiger Wasserstand

const STARK = 10 // 5 - starke Änderung cm / Zeit
const MID = 5    // 3 - Änderung cm / Zeit
const LOW = 5    // 1 - leichte Änderung cm / Zeit

func getTendencyString(diff float64, withUnicode bool) string {
	// zur Tendenz siehe: https://undine.bafg.de/rhein/zustand-aktuell/rhein_akt_WQ.html
	//   Diffzeitraum: 4h
	//   stark: mehr als +/-10 cm Veränderung in 4h
	//   normal_ mehr als +/-5 cm Veränderung in 4h
	//   konstant: bis zu +/- 5cm Veränderung in 4h

	var sb strings.Builder

	if diff > STARK {
		sb.WriteString("stark steigend")
		if withUnicode {
			sb.WriteString(" \u2197")
		}
	} else if diff > MID {
		sb.WriteString("steigend")
		if withUnicode {
			sb.WriteString(" \u2197")
		}
	} else if diff > LOW {
		sb.WriteString("leicht steigend")
		if withUnicode {
			sb.WriteString(" \u2197")
		}
	} else if diff < -STARK {
		sb.WriteString("stark fallend")
		if withUnicode {
			sb.WriteString(" \u2198")
		}
	} else if diff < -MID {
		sb.WriteString("fallend")
		if withUnicode {
			sb.WriteString(" \u2198")
		}
	} else if diff < -LOW {
		sb.WriteString("leicht fallend")
		if withUnicode {
			sb.WriteString(" \u2198")
		}
	} else {
		sb.WriteString("konstant")
		if withUnicode {
			sb.WriteString(" \u27a1")
		}
	}
	return sb.String()
}

func prepareStatusString(current Measurement, trend string) string {

	var sb strings.Builder

	if len(trend) > 0 {
		trend = " - " + trend
	}
	wtime := current.Timestamp.Format("15:04")

	sb.WriteString(fmt.Sprintf("Stand am Pegel Köln um %v Uhr (%v °C): %v cm%v\n", wtime, current.Temperature, current.Level, trend))
	if current.Level >= KATA_02 {
		sb.WriteString("\u26a0 Überflutung der Altstadt \U0001f30a, Rheinufertunnel gesperrt \U0001f6a7\n")
	} else if current.Level >= KATA_01 {
		sb.WriteString("\u26a0 Rheinufertunnel gesperrt \U0001f6a7\n")
	}
	if current.Level >= MARK_02 {
		sb.WriteString("\u26a0 Hochwassermarke 2 erreicht - Schiffsverkehr gesperrt \U0001f6e5\n")
	} else if current.Level >= MARK_01 {
		sb.WriteString("\u26a0 Hochwassermarke 1 erreicht - Schiffsverkehr eingeschränkt \U0001f6e5\n")
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
		fmt.Println("Error encoding timestamp:", err)
		return
	}
	err = os.WriteFile(".kpg_last", val, 0664)
	if err != nil {
		fmt.Println("Error saving file", err)
	}
}

func lastPostTime() (time.Time, error) {
	ts := time.Unix(0, 0)
	data, err := os.ReadFile(".kpg_last")
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("Error reading file:", err)
			return ts, err
		} else {
			return ts, nil
		}
	}
	err = ts.UnmarshalText(data)
	if err != nil {
		fmt.Println("Error decoding timestamp:", err)
		return ts, err
	}
	return ts, nil
}

func checkIfPostNow(measurement Measurement) bool {

	if measurement.Level > KATA_01 {
		// post all available data!
		return true
	}

	lastPost, err := lastPostTime()
	if err != nil {
		return false
	}
	diffMin := int(measurement.Timestamp.Sub(lastPost).Minutes())
	if measurement.Level > MARK_02 {
		// all 30 minutes
		return diffMin >= 30
	} else if measurement.Level > MARK_01 {
		// all 60 minutes
		return diffMin >= 60
	} else {
		// all 3 hours
		return diffMin >= 180
	}
}

func main() {

	readEnvironment("mastodon.env")
	// lst := []string{"SERVER", "CLIENT_ID", "CLIENT_SECRET", "ACCESS_TOKEN"}
	// for _, val := range lst {
	// 	fmt.Printf("%v: '%v'\n", val, os.Getenv(val))
	// }

	// calcTablePosition(22)
	// calcTablePosition(23)
	// calcTablePosition(0)
	// calcTablePosition(1)
	// fmt.Println("22: ", fetchPrognosis(22))
	// fmt.Println("23: ", fetchPrognosis(23))
	// fmt.Println("00: ", fetchPrognosis(0))
	// fmt.Println("01: ", fetchPrognosis(1))

	loadHistory()

	current := retrieveCurrentData()

	var tendency, postTendency string
	diff, err := levelDifference(current, 240)
	if err != nil {
		tendency = ""
		postTendency = ""
	} else {
		tendency = getTendencyString(diff, false)
		postTendency = getTendencyString(diff, true)
	}

	statusText := prepareStatusString(current, postTendency)

	doPost := checkIfPostNow(current)
	if doPost {
		id, err := postToMastodon(statusText)
		if err != nil {
			fmt.Printf("%v [%v] - error: %v\n", current, tendency, err)
		} else {
			fmt.Printf("%v [%v] - %v\n", current, tendency, id)
			savePostTime(current)
		}
	} else {
		fmt.Printf("%v [%v]- no post\n", current, tendency)
	}

	if history[len(history)-1].Timestamp.Before(current.Timestamp) {
		history = append(history, current)
		saveHistory()
	}

	// direction := prognosis - int(waterlevel)
	// trend := "konstant"
	// if direction > 0 {
	// 	trend = "steigend"
	// } else {
	// 	trend = "fallend"
	// }
	// wtime := timestamp.Format("15:04")
	// fmt.Printf("Wasserstand um %v Uhr: %v cm (%v °C) - %v (%v)\n", wtime, waterlevel, watertemp, trend, prognosis)
}
