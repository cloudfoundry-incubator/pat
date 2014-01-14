package pat

import (
	"github.com/julz/pat/benchmarker"
)

type Response struct {
	TotalTime int64
}

func RunCommandLine() *Response {
	totalTime := benchmarker.Time(push)
	return &Response{totalTime.Nanoseconds()}
}
