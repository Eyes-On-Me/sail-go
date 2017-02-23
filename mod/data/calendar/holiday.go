package calendar

import (
	"github.com/sail-services/sail-go/com/data/datetime"
	"time"
)

type ObservedRule int

const (
	ObservedNearest ObservedRule = iota
	ObservedExact
	ObservedMonday
)

var (
	USNewYear           = HolidayNew(time.January, 1)
	USMLK               = HolidayNewFloat(time.January, time.Monday, 3)
	USPresidents        = HolidayNewFloat(time.February, time.Monday, 3)
	USMemorial          = HolidayNewFloat(time.May, time.Monday, -1)
	USMemorialBefore    = HolidayNewFloat(time.May, time.Monday, -2)
	USIndependence      = HolidayNew(time.July, 4)
	USLabor             = HolidayNewFloat(time.September, time.Monday, 1)
	USColumbus          = HolidayNewFloat(time.October, time.Monday, 2)
	USVeterans          = HolidayNew(time.November, 11)
	USThanksgiving      = HolidayNewFloat(time.November, time.Thursday, 4)
	USThanksgivingAfter = HolidayNewFloat(time.November, time.Thursday, 5)
	USChristmas         = HolidayNew(time.December, 25)
	USChristmasBefore   = HolidayNew(time.December, 24)
)

type Holiday struct {
	Month   time.Month
	Weekday time.Weekday
	Day     int
	Offset  int
}

func HolidayNew(month time.Month, day int) Holiday {
	return Holiday{Month: month, Day: day}
}

func HolidayNewFloat(month time.Month, weekday time.Weekday, offset int) Holiday {
	return Holiday{Month: month, Weekday: weekday, Offset: offset}
}

func (h *Holiday) matches(date time.Time) bool {
	if h.Month > 0 {
		if date.Month() != h.Month {
			return false
		}
		if h.Day > 0 {
			return date.Day() == h.Day
		}
		if h.Weekday > 0 && h.Offset != 0 {
			return datetime.IsWeekdayN(date, h.Weekday, h.Offset)
		}
	} else if h.Offset > 0 {
		return date.YearDay() == h.Offset
	}
	return false
}
