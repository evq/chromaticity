package utils

import (
	//	"errors"

	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// Mon Jan 2 15:04:05 MST 2006
const datetimeLayout = "2006-01-02T15:04:05"
const timeLayout = "15:04:05"

type dayOfWeekBitmask uint8

const (
	SUNDAY dayOfWeekBitmask = 1 << iota
	SATURDAY
	FRIDAY
	THURSDAY
	WEDNESDAY
	TUESDAY
	MONDAY
)

var weekdayToDayOfWeek map[time.Weekday]dayOfWeekBitmask = map[time.Weekday]dayOfWeekBitmask{
	time.Sunday:    SUNDAY,
	time.Saturday:  SATURDAY,
	time.Friday:    FRIDAY,
	time.Thursday:  THURSDAY,
	time.Wednesday: WEDNESDAY,
	time.Tuesday:   TUESDAY,
	time.Monday:    MONDAY,
}

var recurringRegex *regexp.Regexp = regexp.MustCompile("W(\\d{1,3})/T(\\d{2}:\\d{2}:\\d{2})")

func (f dayOfWeekBitmask) hasFlag(flag dayOfWeekBitmask) bool {
	return f&flag != 0
}

func (f dayOfWeekBitmask) hasDayOfWeek(day time.Weekday) bool {
	return f.hasFlag(weekdayToDayOfWeek[day])
}

func dayOfWeekBitmaskFrom(bitmask int) (*dayOfWeekBitmask, error) {
	if bitmask == 0 {
		return nil, nil
	}
	if bitmask > 127 {
		return nil, errors.New("Day of week bitmask cannot be > 127")
	}
	mask := dayOfWeekBitmask(bitmask)
	return &mask, nil

}

/**
 * See https://www.developers.meethue.com/documentation/datatypes-and-time-patterns#16_time_patterns
 * for support time patterns.
 */
func GetNextTimeFrom(s string, t *time.Time) (*time.Time, error) {
	if t == nil {
		temp := time.Now()
		t = &temp
	}

	if len(s) == 0 {
		return nil, errors.New("LocalTime must be specified")
	}

	subMatches := recurringRegex.FindStringSubmatch(s)
	if len(subMatches) == 3 {
		// then is a recurring time

		enabledMaskRaw, err := strconv.Atoi(subMatches[1])
		if err != nil {
			return nil, errors.Wrap(err, "Day of week bitmask was not an int")
		}
		enabledMask, err := dayOfWeekBitmaskFrom(enabledMaskRaw)
		if err != nil {
			return nil, err
		}
		if enabledMask == nil {
			return nil, nil
		}
		recurringTime, err := time.Parse(timeLayout, subMatches[2])
		if err != nil {
			return nil, errors.Wrap(err, "Time was not valid")
		}

		hour, min, sec := recurringTime.Clock()
		year, month, day := t.Date()
		nextRecurringTime := time.Date(year, month, day, hour, min, sec, 0, time.UTC)

		// get next future enabled day
		for {
			if nextRecurringTime.After(*t) && enabledMask.hasDayOfWeek(nextRecurringTime.Weekday()) {
				return &nextRecurringTime, nil
			}
			nextRecurringTime = nextRecurringTime.Add(time.Hour * 24)
		}
	}

	// then treat as an absolute time
	absoluteTime, err := time.Parse(datetimeLayout, s)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse as either absolute or recurring time")
	}

	if absoluteTime.Before(*t) {
		return nil, errors.New(fmt.Sprintf("Parsed absolute time %s is before %s", absoluteTime.String(), t.String()))
	}
	return &absoluteTime, err
}
