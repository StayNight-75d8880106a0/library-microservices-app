package helper

import "time"

var jakartaLocation = loadJakartaLocation()

func loadJakartaLocation() *time.Location {
	location, errLocation := time.LoadLocation("Asia/Jakarta")
	if errLocation != nil {
		panic("Failed to load Jakarta location: " + errLocation.Error())
	}

	return location
}

func formatTimeRFC3339Jakarta(t time.Time) string {
	return t.In(jakartaLocation).Format(time.RFC3339)
}

func FormatEpochMillisRFC3339Jakarta(ms int64) string {

	if ms <= 0 {
		return ""
	}

	return formatTimeRFC3339Jakarta(time.UnixMilli(ms))

}
