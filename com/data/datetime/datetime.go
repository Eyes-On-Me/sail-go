package datetime

import (
	"errors"
	"math"
	"time"
)

func EasterDay(year int) (time.Time, error) {
	if year < 326 {
		return time.Now(), errors.New("year have to be greater than 325")
	}
	var a, b, c, d, e int
	var month time.Month
	var day float64
	a = year % 4
	b = year % 7
	c = year % 19
	d = (19*c + 15) % 30
	e = (2*a + 4*b - d + 34) % 7
	month = time.Month((d + e + 114) / 31)
	day = math.Floor(float64((d+e+114)%31 + 1))
	day = day + 13
	return time.Date(year, month, int(day), 0, 0, 0, 0, time.Local), nil
}

func IsWeekend(date time.Time) bool {
	day := date.Weekday()
	return day == time.Saturday || day == time.Sunday
}

func IsWeekdayN(date time.Time, day time.Weekday, n int) bool {
	cday := date.Weekday()
	if cday != day || n == 0 {
		return false
	}
	if n > 0 {
		return (date.Day()-1)/7 == (n - 1)
	} else {
		n = -n
		last := time.Date(date.Year(), date.Month()+1,
			1, 12, 0, 0, 0, date.Location())
		lastCount := 0
		for {
			last = last.AddDate(0, 0, -1)
			if last.Weekday() == day {
				lastCount++
			}
			if lastCount == n || last.Month() != date.Month() {
				break
			}
		}
		return lastCount == n && last.Equal(date)
	}
}
