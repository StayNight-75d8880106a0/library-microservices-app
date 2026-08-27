package helper

import (
	"strings"
	"time"
)

func FormatTimeRFC3339Jakarta(t time.Time) string {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return t.In(loc).Format(time.RFC3339)
}

func NormalizeDate(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	layouts := []string{
		"2006-01-02",      // 2014-01-15 (Standard ISO)
		"January 2, 2006", // October 1, 1988
		"Jan 2, 2006",     // Oct 1, 1988
		"2 January 2006",  // 1 October 1988
		"2 Jan 2006",      // 1 Oct 1988
		"January 2006",    // October 1988
		"Jan 2006",        // Oct 1988
		"2006",            // 1988 (Hanya Tahun)
		"02-01-2006",      // 15-01-2014
		"02/01/2006",      // 15/01/2014
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, input)
		if err == nil {
			if layout == "2006" {
				return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
			}
			return t.Format("2006-01-02")
		}
	}

	return input
}
