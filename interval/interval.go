package interval

import (
	"time"
)

type RepeatItem struct {
	fn       func()
	interval int
	ticker   *time.Ticker
	quit     chan bool
}

var worklist []RepeatItem = make([]RepeatItem, 0)

func Repeat(second int, fn func()) *RepeatItem {
	if second > 0 {
		ri := RepeatItem{fn, second, nil, nil}
		ri.quit, ri.ticker = doWork(second, fn)
		worklist = append(worklist, ri)
		return &ri
	} else {
		fn()
		return &RepeatItem{fn, 0, nil, nil}
	}
}

func (ri *RepeatItem) Stop() {
	ri.quit <- true
}

func doWork(s int, fn func()) (chan bool, *time.Ticker) {
	fn()
	ticker := time.NewTicker(time.Duration(s) * time.Second)
	quit := make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				fn()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	return quit, ticker
}
