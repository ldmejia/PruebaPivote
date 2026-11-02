package shared

import (
	"time"
	_ "time/tzdata"
)

var Guatemala = loadGuatemala()

func loadGuatemala() *time.Location {
	loc, err := time.LoadLocation("America/Guatemala")
	if err != nil {
		return time.FixedZone("CST", -6*60*60)
	}
	return loc
}

func LocalDay(t time.Time) time.Time {
	y, m, d := t.In(Guatemala).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
