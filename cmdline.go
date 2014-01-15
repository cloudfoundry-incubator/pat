package pat

import (
	"fmt"
	. "github.com/julz/pat/benchmarker"
	"time"
)

type Response struct {
	TotalTime int64
	Timestamp int64
}

func RunCommandLine(pushes int, concurrency int) *Response {
	result := make(chan time.Duration)
	go func(result chan time.Duration) {
		var avg int64
		var total int64
		var n int64
		for r := range result {
			total = total + r.Nanoseconds()
			n = n + 1
			avg = total / n
			fmt.Print("\033[2J\033[;H")
			fmt.Println("Cloud Foundry Performance Acceptance Tests")
			fmt.Printf("Test underway.  Pushes: \x1b[32m%v\x1b[0m  Concurrency: \x1b[32m%v\x1b[0m\n", pushes, concurrency)
			fmt.Println("----------------------------------------------------------\n")
			fmt.Printf("Total pushes:   \x1b[32m%v\x1b[0m\n", n)
			fmt.Printf("Latest Push:    \x1b[32m%v\x1b[0m\n", r.Nanoseconds())
			fmt.Printf("Average:        \x1b[32m%v\x1b[0m\n", avg)
			fmt.Printf("Total time:     \x1b[32m%v\x1b[0m\n", total)
		}
	}(result)

	totalTime := Time(func() { ExecuteConcurrently(concurrency, Repeat(pushes, Timed(result, dummy))) })
	return &Response{totalTime.Nanoseconds(), time.Now().UnixNano()}
}
