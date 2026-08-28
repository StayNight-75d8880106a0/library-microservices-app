package helper

import (
	"time"
)

func FormatTimeRFC3339Jakarta(t time.Time) string {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return t.In(loc).Format(time.RFC3339)
}
