package modules

import (
	"github.com/revel/revel"
)

type RequestLog struct {
	Queries   []QueryLog
	User      string
	URL       string
	TimeTaken int64
	QueryTime int64
}

type QueryLog struct {
	SQL       string
	Params    interface{}
	TimeTaken int64
	ShortSQL  string
}

var Requests map[*revel.Controller]*RequestLog

var RequestLogs map[string]*PathLog

// PathLog keeps a running total of requests to a path and stores details for
// the most recent of these.
//
// `RequestLogs` is expected to be used as a rotating log (queue), where
// `MostRecent` holds the index of the most recent request added to
// `RequestLogs`.
type PathLog struct {
	// `MostRecent` has type of `int` because this is the return type of `cap`,
	// which defines the largest size of a go array/slice.
	MostRecent  int
	RequestLogs []*RequestLog
	NumCalls    int64
}

func NewPathLog(logSize int) *PathLog {
	return &PathLog{
		MostRecent:  -1,
		RequestLogs: make([]*RequestLog, 0, logSize),
		NumCalls:    0,
	}
}
