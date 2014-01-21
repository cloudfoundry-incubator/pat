package interval 

import (
	"fmt"
	"time"
)

type RepeatItem struct {
	fn func()
	interval int
	ticker *time.Ticker
	quit chan bool
}

var worklist []RepeatItem = make([]RepeatItem, 0)

func Repeat(second int, fn func()) *RepeatItem {
	if (second > 0) {
		w := RepeatItem{ fn, second, nil, nil }
		w.quit, w.ticker = doWork(second, fn)
		worklist = append(worklist, w)
		return &w
	} else {
		fn()
		return nil
	}
}

func (w *RepeatItem) Stop() {
	w.quit <- true
}

func doWork(s int, fn func()) (chan bool, *time.Ticker) {
	ticker := time.NewTicker(time.Duration(s) * time.Second)
	quit := make(chan bool)
	go func() {
		defer fmt.Println("job stopped")
		for {
			select {
				case <- ticker.C:
					fn()
				case <- quit:
					ticker.Stop()
				return
			}
		}
	}()
	return quit, ticker
}


