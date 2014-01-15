package pat

import (
	"github.com/julz/pat/benchmarker"
	"time"
)

type Response struct {
	TotalTime int64
	Timestamp int64
}

func RunCommandLine() *Response {
	totalTime := benchmarker.Time(push)
	return &Response{totalTime.Nanoseconds(), time.Now().UnixNano()}
}
