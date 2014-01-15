package main

import (
	"flag"
	"fmt"
	"github.com/julz/pat"
)

func main() {
	server := flag.Bool("server", false, "true to run the HTTP server interface")
	flag.Parse()

	if *server == true {
		fmt.Println("Starting in server mode")
		pat.Serve()
		pat.Bind()
	} else {
		resp := pat.RunCommandLine()
		fmt.Printf("Total Time: %d", resp.TotalTime)
	}
}
