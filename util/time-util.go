package util

import (
	"time"
)

//ToTime Converts string time to time.Time
func ToTime(strTime string) *time.Time {
	t, err := time.Parse(time.RFC3339, strTime)
	if err != nil {
		t, _ = time.Parse("2006-01-02T15:04:05.000Z07:00", strTime)
	}
	return &t
}
