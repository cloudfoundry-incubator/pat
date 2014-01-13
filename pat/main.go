package main

import (
  "fmt"
  "github.com/julz/pat"
)

func main() {
  resp := pat.RunCommandLine()
  fmt.Printf("Total Time: %d", resp.TotalTime)

  //  pat.Serve()
  //  pat.Bind()
}
