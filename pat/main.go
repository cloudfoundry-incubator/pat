package main

import (
  "flag"
  "fmt"
  "github.com/julz/pat"
  "github.com/julz/pat/server"
)

func main() {
  useServer := flag.Bool("server", false, "true to run the HTTP server interface")
  pushes := flag.Int("pushes", 1, "number of pushes to attempt")
  concurrency := flag.Int("concurrency", 1, "max number of pushes to attempt in parallel")
  silent := flag.Bool("silent", false, "true to run the commands and print output the terminal")
  output := flag.String("output", "", "if specified, writes benchmark results to a CSV file")
  flag.Parse()

  if *useServer == true {
    fmt.Println("Starting in server mode")
    server.Serve()
    server.Bind()
  } else {
    pat.RunCommandLine(*pushes, *concurrency, *silent, *output)
  }
}
