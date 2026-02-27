package timeutil

import (
	"time"
)

// BangladeshLocation is the Asia/Dhaka timezone (UTC+6).
var BangladeshLocation *time.Location

func init() {
	var err error
	BangladeshLocation, err = time.LoadLocation("Asia/Dhaka")
	if err != nil {
		// Fallback to fixed offset UTC+6
		BangladeshLocation = time.FixedZone("BST", 6*60*60)
	}
}

// NowBD returns the current time in Bangladesh timezone.
func NowBD() time.Time {
	return time.Now().In(BangladeshLocation)
}

// ToBD converts a time to Bangladesh timezone.
func ToBD(t time.Time) time.Time {
	return t.In(BangladeshLocation)
}

// ToUTC converts a time to UTC.
func ToUTC(t time.Time) time.Time {
	return t.UTC()
}

// StartOfDayBD returns the start of the day (00:00:00) in Bangladesh timezone.
func StartOfDayBD(t time.Time) time.Time {
	bd := ToBD(t)
	return time.Date(bd.Year(), bd.Month(), bd.Day(), 0, 0, 0, 0, BangladeshLocation)
}

// EndOfDayBD returns the end of the day (23:59:59.999999999) in Bangladesh timezone.
func EndOfDayBD(t time.Time) time.Time {
	bd := ToBD(t)
	return time.Date(bd.Year(), bd.Month(), bd.Day(), 23, 59, 59, 999999999, BangladeshLocation)
}

// OperationalDate returns the "business date" in Bangladesh timezone.
// For orders placed after midnight but before 5 AM, this returns the previous day.
func OperationalDate(t time.Time) time.Time {
	bd := ToBD(t)
	if bd.Hour() < 5 {
		bd = bd.AddDate(0, 0, -1)
	}
	return time.Date(bd.Year(), bd.Month(), bd.Day(), 0, 0, 0, 0, BangladeshLocation)
}

// FormatBD formats a time in Bangladesh timezone using the given layout.
func FormatBD(t time.Time, layout string) string {
	return ToBD(t).Format(layout)
}

// ParseBD parses a time string in Bangladesh timezone.
func ParseBD(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, BangladeshLocation)
}
