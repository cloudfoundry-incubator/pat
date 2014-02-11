package main

import (
	"fmt"
	"github.com/julz/pat"
	"github.com/julz/pat/config"
	"github.com/julz/pat/server"
)

func main() {
	config := config.NewConfig()
	err := config.Parse()
	if err != nil {
		panic(err)
	}

	if config.Server == true {
		fmt.Println("Starting in server mode")
		server.Serve()
		server.Bind()
	} else {
		pat.RunCommandLine(config)
	}
}
