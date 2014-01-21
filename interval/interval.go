package main 

import (
	"fmt"
	"time"
)

type workitem struct {
	fn func()
	interval int
	ticker *time.Ticker
	quit chan bool
}

var worklist []workitem = make([]workitem, 0)

func main() {
	goRepeat( 2, func(){ fmt.Printf("%v Time in 2\n", time.Stamp) } )
	time.AfterFunc(10 * time.Second, func(){
		worklist[0].quit <- true
	})	

var quit = make(chan bool)
<-quit
}

func goRepeat(s int, fn func()) {
	w :=  workitem{fn, s, nil, nil}
	w.quit, w.ticker = doWork(s, fn)
	worklist = append(worklist, w)
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

