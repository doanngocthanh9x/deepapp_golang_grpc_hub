package utils

import "time"

func Now() time.Time {
	return time.Now()
}

func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}