package utils

import (
	"fmt"
	"testing"
	"time"
)

/**
 * [YYYY]-[MM]-[DD]T[hh]:[mm]:[ss]
 * ([date]T[time])
 */
func TestAbsoluteTime(t *testing.T) {
	t1 := "2018-03-03T01:02:03"
	et1 := time.Date(2018, time.March, 3, 1, 2, 3, 0, time.UTC)
	t2 := time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC)
	localTime, err := GetNextTimeFrom(t1, &t2)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !localTime.Equal(et1) {
		t.Errorf(fmt.Sprintf("%s not equal to %s", localTime.String(), t1))
	}
}

/**
 * W[bbb]/T[hh]:[mm]:[ss]
 * Every day of the week given by bbb at given time
 * (W/[time])
 * so bbb = 0MTWTFSS – So only Tuesdays is 00100000 = 32
 */
func TestRecurringTimeEveryDay(t *testing.T) {
	t1 := "W127/T01:02:03"
	et1 := time.Date(2018, time.January, 1, 1, 2, 3, 0, time.UTC)
	t2 := time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC)
	localTime, err := GetNextTimeFrom(t1, &t2)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !localTime.Equal(et1) {
		t.Errorf(fmt.Sprintf("%s not equal to %s", localTime.String(), et1))
	}
}

func TestRecurringTimeMonday(t *testing.T) {
	t1 := "W64/T01:02:03"
	et1 := time.Date(2018, time.January, 1, 1, 2, 3, 0, time.UTC)
	t2 := time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC) // Monday
	//fmt.Printf(t2.Weekday().String())
	localTime, err := GetNextTimeFrom(t1, &t2)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !localTime.Equal(et1) {
		t.Errorf(fmt.Sprintf("%s not equal to %s", localTime.String(), et1))
	}
}

func TestRecurringTimeNextMonday(t *testing.T) {
	t1 := "W64/T01:02:03"
	et1 := time.Date(2018, time.January, 8, 1, 2, 3, 0, time.UTC)
	t2 := time.Date(2018, time.January, 2, 0, 0, 0, 0, time.UTC) // Monday
	fmt.Printf(t2.Weekday().String())
	localTime, err := GetNextTimeFrom(t1, &t2)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !localTime.Equal(et1) {
		t.Errorf(fmt.Sprintf("%s not equal to %s", localTime.String(), et1))
	}
}

func TestRecurringTimeInvalidBitmask(t *testing.T) {
	t1 := "W129/T01:02:03"
	t2 := time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC)
	_, err := GetNextTimeFrom(t1, &t2)
	if err == nil {
		t.Errorf("Error expected, but no error")
	}
}
