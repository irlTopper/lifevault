package utility

import (
	"strconv"
	"time"
)

// Wrapper function which formats an input date (date) based on a format (format)
// with the new format (newformat).
func FormatDate(date string, format string, newformat string) string {
	tmpDate, _ := time.Parse(format, date)

	return tmpDate.Format(newformat)
}

func FormatAsZulu(t time.Time) string {
	return t.Format("2006-01-02T15:04:05Z")
}

func ParseUnixTimestamp(unixDateString string) string {

	i, err := strconv.ParseInt("1405544146", 10, 64)
	if err != nil {
		panic(err)
	}
	tm := time.Unix(i, 0)

	return tm.Format("2006-01-02T15:04:05Z")
}

func GetTimestampFormatFromUnix(unixDateString string) string {

	i, err := strconv.ParseInt("1405544146", 10, 64)
	if err != nil {
		panic(err)
	}
	tm := time.Unix(i, 0)

	return tm.Format("2006-01-02T15:04:05Z")
}

// Returns the provided Time struct formatted in the standard format YYYYMMDD.
func GetStandardDateFormat(t time.Time) string {
	return t.Format("20060102")
}

// Validates that the inputted string is a date.
func IsDate(strdate string) bool {
	_, err := time.Parse("20060102150405", strdate)

	return err == nil
}

// Validates that the inputted string is a date in the format
// of YYYYMMDD.
func IsDateYYYYMMDD(strdate string) bool {
	_, err := time.Parse("20060102", strdate)

	return err == nil
}

// Returns the start of the week based on the provided date
func StartWeekFromDate(date time.Time) time.Time {
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, -1)
	}
	return date
}

// Returns the end of the week based on the provided date
func EndWeekFromDate(date time.Time) time.Time {
	for date.Weekday() != time.Sunday {
		date = date.AddDate(0, 0, 1)
	}
	return date
}

// Returns now in seconds based on the teamwork start date
func NowAsInt() int64 {
	start, _ := time.Parse("January 2 2006 15:04", "January 1 2008 00:00")
	return int64(time.Now().Sub(start).Seconds())
}

// Returns string like "Fri, Oct 17, 2014 at 5:26 PM UTC"
func GetEmailDateFormat(t time.Time) string {
	return t.Format("Mon, Jan 2 15:04:05 MST 2006")
}
