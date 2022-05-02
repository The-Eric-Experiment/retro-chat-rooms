package helpers

import "time"

func Now() time.Time {
	return time.Now().UTC()
}

func FormatTimestamp(t time.Time) string {
	return t.Format("03:04:05 PM")
}
