package main

import (
	"os"
	"flag"
	"fmt"
	"github.com/julz/pat"
	"github.com/julz/pat/server"
	"github.com/julz/pat/parser"
)

func main() {
	useServer := flag.Bool("server", false, "true to run the HTTP server interface")
	pushes := flag.Int("pushes", 1, "number of pushes to attempt")
	concurrency := flag.Int("concurrency", 1, "max number of pushes to attempt in parallel")
	silent := flag.Bool("silent", false, "true to run the commands and print output the terminal")
	output := flag.String("output", "", "if specified, writes benchmark results to a CSV file")
	config := flag.String("config", "", "name of the command line configuration file you wish to use (must be saved under the config directory)")
	flag.Parse()

	fmt.Println(os.Args[1:])

	if *config != "" {
		cfg, err := parser.NewPATsConfiguration(*config)
		if err != nil {
			panic(err) //(dan) should just report the error and stop if there is an error loading the configuration file
		}
		useServer = cfg.Cli_command.Server
		pushes = cfg.Cli_command.Pushes
		concurrency = cfg.Cli_commands.Concurrency
		silent = cfg.Cli_commands.Silent
		output = cfg.Cli_commands.Output
	}
	if *useServer == true {
		fmt.Println("Starting in server mode")
		server.Serve()
		server.Bind()
	} else {
		pat.RunCommandLine(*pushes, *concurrency, *silent, *output)
	}
}
