package pat

import (
	"fmt"
	. "github.com/julz/pat/benchmarker"
	"strings"
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
			fmt.Printf("Total pushes:   \x1b[32m%v\x1b[0m / %v           %v\n", n, int64(pushes), bar(n, int64(pushes), 20))
			fmt.Printf("Latest Push:    \x1b[32m%v\x1b[0m\n", r.Nanoseconds())
			fmt.Printf("Average:        \x1b[32m%v\x1b[0m\n", avg)
			fmt.Printf("Total time:     \x1b[32m%v\x1b[0m\n", total)
			fmt.Println("----------------------------------------------------------\n")

		}
	}(result)

	totalTime := Time(func() { ExecuteConcurrently(concurrency, Repeat(pushes, Timed(result, dummy))) })
	return &Response{totalTime.Nanoseconds(), time.Now().UnixNano()}
}

func bar(n int64, total int64, size int) (bar string) {
	progress := int64(size) / (total / n)
	return "|" + strings.Repeat("X", int(progress)) + strings.Repeat("-", size-int(progress)) + "|"
}
