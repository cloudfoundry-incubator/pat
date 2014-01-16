package pat

import (
  "fmt"
  "net/http"
)

func CallPush(w http.ResponseWriter, pushes int, concurrency int) {
  resp := RunCommandLine(pushes, concurrency)
  fmt.Fprintf(w, "\n\nTest Complete. Total Time: %d", resp.TotalTime)
}
