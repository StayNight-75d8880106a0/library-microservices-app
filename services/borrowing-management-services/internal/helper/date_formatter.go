package helper

import "time"

var jakartaLoc = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*60*60)
	}
	return loc
}()

func FormatTimeRFC3339Jakarta(t time.Time) string {
	return t.In(jakartaLoc).Format(time.RFC3339)
}

func FormatTimeRFC3339JakartaPTR(t *time.Time) string {
	if t == nil {
		return ""
	}
	return FormatTimeRFC3339Jakarta(*t)
}
