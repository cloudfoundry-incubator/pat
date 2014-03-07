package main

import (
	"fmt"
	"os"

	"github.com/julz/pat/cmdline"
	"github.com/julz/pat/config"
	"github.com/julz/pat/server"
)

func main() {
	useServer := false
	flags := config.ConfigAndFlags
	flags.BoolVar(&useServer, "server", false, "true to run the HTTP server interface")

	cmdline.InitCommandLineFlags(flags)
	flags.Parse(os.Args[1:])

	if useServer == true {
		fmt.Println("Starting in server mode")
		server.Serve()
		server.Bind()
	} else {
		cmdline.RunCommandLine()
	}
}
