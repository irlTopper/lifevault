package utility

import "time"

func GetEpochMilliseconds() int64 {
	return time.Now().UnixNano() / 1000000
}
