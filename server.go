package pat

import (
  "fmt"
  "github.com/julz/pat/benchmarker"
  "net/http"
)

func Serve() {
  http.HandleFunc("/experiments/push", handlePush)
}

func Stop() {
}

func Bind() {
  http.ListenAndServe(":8080", nil)
}

func handlePush(w http.ResponseWriter, r *http.Request) {
  totalTime := benchmarker.Time(push)
  fmt.Fprintf(w, "{ \"totalTime\": %d }", totalTime.Nanoseconds())
}
