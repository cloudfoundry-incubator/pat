package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/pat/cmdline"
	"github.com/cloudfoundry-incubator/pat/config"
	"github.com/cloudfoundry-incubator/pat/logs"
	"github.com/cloudfoundry-incubator/pat/server"
)

func main() {
	useServer := false
	flags := config.ConfigAndFlags
	flags.BoolVar(&useServer, "server", false, "true to run the HTTP server interface")

	logs.InitCommandLineFlags(flags)
	cmdline.InitCommandLineFlags(flags)
	server.InitCommandLineFlags(flags)
	flags.Parse(os.Args[1:])

	if useServer == true {
		logs.NewLogger("main").Info("Starting in server mode")
		server.Serve()
	} else {
		err := cmdline.RunCommandLine()
		if err != nil {
			fmt.Println(err)
		}
	}
}
