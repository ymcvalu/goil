package helper

import (
	"time"
)

const (
	Nanosecond  int64 = 1
	Microsecond       = 1000 * Nanosecond
	Millisecond       = 1000 * Microsecond
	Second            = 1000 * Millisecond
	Minute            = 60 * Second
	Hour              = 60 * Minute
)

func SecToDuration(sec int64) time.Duration {
	return time.Duration(Second * sec)
}

func MinToDuration(min int64) time.Duration {
	return time.Duration(min * Minute)
}

func HourToDuration(hour int64) time.Duration {
	return time.Duration(hour * Hour)
}

func TimeToDuration(hour, min, sec int64) time.Duration {
	return time.Duration(sec*Second + min*Minute + hour*Hour)
}

func DurationToSecond(d time.Duration) int64 {
	return int64(d) / Second
}

func DurationToMin(d time.Duration) int64 {
	return int64(d) / Minute
}

func DurationToHour(d time.Duration) int64 {
	return int64(d) / Hour
}

func DurationToTime(d time.Duration) (hour, min, sec int64) {
	dd := int64(d)
	hour = dd / Hour
	dd %= Hour
	min = dd / Minute
	dd %= Minute
	sec = dd / Second
	return
}
