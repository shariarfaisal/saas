package timeutil

import (
	"testing"
	"time"
)

func TestNowBD(t *testing.T) {
	now := NowBD()
	if now.Location().String() != BangladeshLocation.String() {
		t.Errorf("expected location %s, got %s", BangladeshLocation, now.Location())
	}
}

func TestToBD(t *testing.T) {
	utc := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	bd := ToBD(utc)

	// UTC+6, so 12:00 UTC = 18:00 BD
	if bd.Hour() != 18 {
		t.Errorf("expected hour 18 in BD, got %d", bd.Hour())
	}
}

func TestStartOfDayBD(t *testing.T) {
	utc := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	start := StartOfDayBD(utc)

	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Errorf("expected 00:00:00, got %02d:%02d:%02d", start.Hour(), start.Minute(), start.Second())
	}
}

func TestEndOfDayBD(t *testing.T) {
	utc := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
	end := EndOfDayBD(utc)

	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Errorf("expected 23:59:59, got %02d:%02d:%02d", end.Hour(), end.Minute(), end.Second())
	}
}

func TestOperationalDate_AfterMidnight(t *testing.T) {
	// 2 AM BD time → should return previous day
	bd2am := time.Date(2024, 6, 15, 2, 0, 0, 0, BangladeshLocation)
	opDate := OperationalDate(bd2am)

	if opDate.Day() != 14 {
		t.Errorf("expected operational date day 14, got %d", opDate.Day())
	}
}

func TestOperationalDate_AfterFiveAM(t *testing.T) {
	// 6 AM BD time → should return same day
	bd6am := time.Date(2024, 6, 15, 6, 0, 0, 0, BangladeshLocation)
	opDate := OperationalDate(bd6am)

	if opDate.Day() != 15 {
		t.Errorf("expected operational date day 15, got %d", opDate.Day())
	}
}

func TestParseBD(t *testing.T) {
	parsed, err := ParseBD("2006-01-02 15:04", "2024-06-15 18:00")
	if err != nil {
		t.Fatalf("ParseBD failed: %v", err)
	}

	if parsed.Location().String() != BangladeshLocation.String() {
		t.Errorf("expected BD location, got %s", parsed.Location())
	}
}

func TestFormatBD(t *testing.T) {
	utc := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	formatted := FormatBD(utc, "15:04")

	if formatted != "18:00" {
		t.Errorf("expected 18:00, got %s", formatted)
	}
}
