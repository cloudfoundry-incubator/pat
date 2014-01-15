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
	errors := make(chan error)
	workers := make(chan int)
	go func(result chan time.Duration) {
		var avg int64
		var total int64
		var n int64
		var totalErrors int64
		var lastError string
		var lastPush int64
		var worstPush int64
		var running int
		for {
			select {
			case r := <-result:
				total = total + r.Nanoseconds()
				n = n + 1
				avg = total / n
				lastPush = r.Nanoseconds()
				if lastPush > worstPush {
					worstPush = lastPush
				}
			case w := <-workers:
				running = running + w
			case e := <-errors:
				totalErrors = totalErrors + 1
				lastError = e.Error()
			}

			fmt.Print("\033[2J\033[;H")
			fmt.Println("Cloud Foundry Performance Acceptance Tests")
			fmt.Printf("Test underway.  Pushes: \x1b[32m%v\x1b[0m  Concurrency: \x1b[32m%v\x1b[0m\n", pushes, concurrency)
			fmt.Println("----------------------------------------------------------\n")
			fmt.Printf("Total pushes:    \x1b[32m%v\x1b[0m / %v        %v\n", n, int64(pushes), bar(n, int64(pushes), 25))
			fmt.Printf("Latest Push:     \x1b[32m%v\x1b[0m\n", lastPush)
			fmt.Printf("Worst Push:      \x1b[32m%v\x1b[0m\n", worstPush)
			fmt.Printf("Average:         \x1b[32m%v\x1b[0m\n", avg)
			fmt.Printf("Total time:      \x1b[32m%v\x1b[0m\n", total)
			fmt.Printf("Running Workers: \x1b[32m%v\x1b[0m\n", running)
			fmt.Println("----------------------------------------------------------\n")
			if totalErrors > 0 {
				fmt.Printf("Total errors: %d\n", totalErrors)
				fmt.Printf("Last error: %v\n", lastError)
			}
		}
	}(result)

	totalTime, _ := Time(func() error {
		ExecuteConcurrently(concurrency, Repeat(pushes, Counted(workers, Timed(result, errors, push))))
		return nil
	})

	return &Response{totalTime.Nanoseconds(), time.Now().UnixNano()}
}

func bar(n int64, total int64, size int) (bar string) {
	if n > 0 {
		progress := int64(size) / (total / n)
		return "|" + strings.Repeat("X", int(progress)) + strings.Repeat("-", size-int(progress)) + "|"
	} else {
		return ""
	}
}
