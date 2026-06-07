package journal

import (
	"testing"
	"time"
)

// TestGetCreationTime_BasicBehavior verifies that getCreationTime:
// - interprets the calendar date in Asia/Kolkata,
// - uses the current local time-of-day from Asia/Kolkata,
// - and returns a UTC instant.
func TestGetCreationTime_BasicBehavior(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Skip("Asia/Kolkata location not available")
	}

	// Simulate a DATE value coming from DB. We use a UTC midnight representation
	// because DATE columns are scanned into time.Time with a zero time component.
	logDateUTC := time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC)

	// Capture current local time components for comparison (deterministic within test run)
	nowLocal := time.Now().In(loc)

	created := getCreationTime(logDateUTC)

	// getCreationTime should return a canonical UTC instant
	if created.Location().String() != "UTC" {
		t.Fatalf("expected created time in UTC but got location=%v", created.Location())
	}

	// When converted to Asia/Kolkata, the calendar date must equal the log date interpreted in Asia/Kolkata
	createdLocal := created.In(loc)
	expectedDate := logDateUTC.In(loc)
	y, m, d := expectedDate.Date()
	cy, cm, cd := createdLocal.Date()
	if y != cy || m != cm || d != cd {
		t.Fatalf("date mismatch: expected %04d-%02d-%02d got %04d-%02d-%02d", y, m, d, cy, cm, cd)
	}

	// The time-of-day should be taken from the current local time. Allow a small tolerance
	// for the time taken between reading nowLocal and executing getCreationTime.
	targetLocal := time.Date(createdLocal.Year(), createdLocal.Month(), createdLocal.Day(), nowLocal.Hour(), nowLocal.Minute(), nowLocal.Second(), nowLocal.Nanosecond(), loc)
	delta := createdLocal.Sub(targetLocal)
	if delta < 0 {
		delta = -delta
	}
	if delta > 2*time.Second {
		t.Fatalf("time-of-day mismatch: expected approx %02d:%02d:%02d, got %02d:%02d:%02d (delta=%v)", nowLocal.Hour(), nowLocal.Minute(), nowLocal.Second(), createdLocal.Hour(), createdLocal.Minute(), createdLocal.Second(), delta)
	}
}

// TestGetCreationTime_DifferentInputLocation ensures getCreationTime treats the
// calendar date consistently even if the incoming time.Time has a different location.
func TestGetCreationTime_DifferentInputLocation(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		t.Skip("Asia/Kolkata location not available")
	}

	// Create a logDate that is midnight in a different location (UTC) but represents the same calendar day
	logDateOther := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)

	nowLocal := time.Now().In(loc)
	created := getCreationTime(logDateOther)
	createdLocal := created.In(loc)

	// Ensure the created date equals the calendar date of logDateOther when interpreted in Asia/Kolkata
	expectedDate := logDateOther.In(loc)
	y, m, d := expectedDate.Date()
	cy, cm, cd := createdLocal.Date()
	if y != cy || m != cm || d != cd {
		t.Fatalf("date mismatch for different input location: expected %04d-%02d-%02d got %04d-%02d-%02d", y, m, d, cy, cm, cd)
	}

	// Also validate time-of-day is taken from local time (allow small tolerance)
	targetLocal := time.Date(createdLocal.Year(), createdLocal.Month(), createdLocal.Day(), nowLocal.Hour(), nowLocal.Minute(), nowLocal.Second(), nowLocal.Nanosecond(), loc)
	delta := createdLocal.Sub(targetLocal)
	if delta < 0 {
		delta = -delta
	}
	if delta > 2*time.Second {
		t.Fatalf("time-of-day mismatch for different input location: expected approx %02d:%02d:%02d, got %02d:%02d:%02d (delta=%v)", nowLocal.Hour(), nowLocal.Minute(), nowLocal.Second(), createdLocal.Hour(), createdLocal.Minute(), createdLocal.Second(), delta)
	}
}
