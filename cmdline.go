package pat

import (
	"encoding/csv"
	"fmt"
	"github.com/julz/pat/experiment"
	"os"
	"strconv"
	"strings"
)

type Response struct {
	TotalTime int64
	Timestamp int64
}

func RunCommandLine(pushes int, concurrency int, silent bool, output string) error {
	return experiment.Run(pushes, concurrency, func(samples chan *experiment.Sample, target int) {
		var w *csv.Writer
		if len(output) > 0 {
			f, _ := os.Create(output)
			w = csv.NewWriter(f)
			w.Write([]string{"duration", "wallTime", "average", "workers"})
			defer func() {
				f.Close()
			}()
		}

		csvRecords := 0
		for s := range samples {
			if len(output) > 0 {
				if s.Type == experiment.ResultSample {
					w.Write([]string{strconv.Itoa(int(s.LastResult.Nanoseconds())), strconv.Itoa(int(s.WallTime.UnixNano())), strconv.Itoa(int(s.Average.Nanoseconds())), strconv.Itoa(int(s.TotalWorkers))})
					w.Flush()
					csvRecords++
				}
			}
			if !silent {
				fmt.Print("\033[2J\033[;H")
				fmt.Println("\x1b[32;1mCloud Foundry Performance Acceptance Tests\x1b[0m")
				fmt.Printf("Test underway.  Pushes: \x1b[36m%v\x1b[0m  Concurrency: \x1b[36m%v\x1b[0m\n", pushes, concurrency)
				fmt.Println("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n")
				fmt.Printf("\x1b[36mTotal pushes\x1b[0m:    %v  \x1b[36m%v\x1b[0m / %v\n", bar(s.Total, int64(target), 25), s.Total, int64(target))
				fmt.Println()
				fmt.Printf("\x1b[1mLatest Push\x1b[0m:     \x1b[36m%v\x1b[0m\n", s.LastResult)
				fmt.Printf("\x1b[1mWorst Push\x1b[0m:      \x1b[36m%v\x1b[0m\n", s.WorstResult)
				fmt.Printf("\x1b[1mAverage\x1b[0m:         \x1b[36m%v\x1b[0m\n", s.Average)
				fmt.Printf("\x1b[1mTotal time\x1b[0m:      \x1b[36m%v\x1b[0m\n", s.TotalTime)
				fmt.Printf("\x1b[1mWall time\x1b[0m:       \x1b[36m%v\x1b[0m\n", s.WallTime)
				fmt.Printf("\x1b[1mRunning Workers\x1b[0m: \x1b[36m%v\x1b[0m\n", s.TotalWorkers)
				fmt.Println()
				fmt.Println("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄")
				if len(output) > 0 {
					fmt.Printf("Writing to CSV: %v (Written %d records)", output, csvRecords)
				}
				if s.TotalErrors > 0 {
					fmt.Printf("\nTotal errors: %d\n", s.TotalErrors)
					fmt.Printf("Last error: %v\n", "")
				}
			}
		}
	})
}

func bar(n int64, total int64, size int) (bar string) {
	if n == 0 {
		n = 1
	}
	progress := int64(size) / (total / n)
	return "╞" + strings.Repeat("═", int(progress)) + strings.Repeat("┄", size-int(progress)) + "╡"
}
