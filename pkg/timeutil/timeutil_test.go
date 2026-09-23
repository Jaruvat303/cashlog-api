package timeutil_test

import (
	"testing"
	"time"

	"github.com/Jaruvat303/cashlog/pkg/timeutil"
	"github.com/stretchr/testify/assert"
)

func TestMonthRangeBangkok(t *testing.T) {
	t.Run("00:30 Bangkok on the 1st falls within the month range", func(t *testing.T) {
		start, end := timeutil.MonthRangeBangkok(2026, 6)
		tx := time.Date(2026, time.June, 1, 0, 30, 0, 0, timeutil.BangKokLoc)

		assert.True(t, !tx.Before(start) && tx.Before(end))
	})

	t.Run("23:30 Bangkok on the last day falls within the month range", func(t *testing.T) {
		start, end := timeutil.MonthRangeBangkok(2026, 6)
		tx := time.Date(2026, time.June, 30, 23, 30, 0, 0, timeutil.BangKokLoc)

		assert.True(t, !tx.Before(start) && tx.Before(end))
	})

	t.Run("instant exactly at end belongs to the next month, not this one (half-open)", func(t *testing.T) {
		_, end := timeutil.MonthRangeBangkok(2026, 6)
		nextStart, _ := timeutil.MonthRangeBangkok(2026, 7)

		assert.Equal(t, nextStart, end)
		assert.False(t, end.Before(end)) // tx.Before(end) is the inclusion test; at tx == end it must be false
	})

	t.Run("December rolls over into January of the next year", func(t *testing.T) {
		start, end := timeutil.MonthRangeBangkok(2026, 12)

		assert.Equal(t, time.Date(2026, time.December, 1, 0, 0, 0, 0, timeutil.BangKokLoc), start)
		assert.Equal(t, time.Date(2027, time.January, 1, 0, 0, 0, 0, timeutil.BangKokLoc), end)
	})

	t.Run("start and end are genuinely Asia/Bangkok, not just offset-equivalent", func(t *testing.T) {
		start, end := timeutil.MonthRangeBangkok(2026, 6)

		assert.Equal(t, "Asia/Bangkok", start.Location().String())
		assert.Equal(t, "Asia/Bangkok", end.Location().String())
	})
}

func TestYearRangeBangkok(t *testing.T) {
	t.Run("start is Jan 1 00:00 Bangkok and end is Jan 1 00:00 Bangkok of the next year", func(t *testing.T) {
		start, end := timeutil.YearRangeBangkok(2026)

		assert.Equal(t, time.Date(2026, time.January, 1, 0, 0, 0, 0, timeutil.BangKokLoc), start)
		assert.Equal(t, time.Date(2027, time.January, 1, 0, 0, 0, 0, timeutil.BangKokLoc), end)
	})
}
