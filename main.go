package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry-community/pat/cmdline"
	"github.com/cloudfoundry-community/pat/config"
	"github.com/cloudfoundry-community/pat/server"
)

func main() {
	useServer := false
	flags := config.ConfigAndFlags
	flags.BoolVar(&useServer, "server", false, "true to run the HTTP server interface")

	logs.InitCommandLingFlags(flags)
	cmdline.InitCommandLineFlags(flags)
	server.InitCommandLineFlags(flags)
	flags.Parse(os.Args[1:])

	if useServer == true {
		fmt.Println("Starting in server mode")
		server.Serve()
	} else {
		err := cmdline.RunCommandLine()
		if err != nil {
			fmt.Println(err)
		}
	}
}
