package config

import (
	"flag"
	"io/ioutil"

	"launchpad.net/goyaml"
)

type Config struct {
	Config      string
	Server      bool
	Iterations  int
	Concurrency int
	Silent      bool
	Output      string
	Workload    string
	Interval    int
	Stop        int
}

func NewConfig() *Config {
	var config Config

	flag.StringVar(&config.Config, "config", "", "name of the command line configuration file you wish to use (including path to file)")
	flag.BoolVar(&config.Server, "server", false, "true to run the HTTP server interface")
	flag.IntVar(&config.Iterations, "iterations", 1, "number of pushes to attempt")
	flag.IntVar(&config.Concurrency, "concurrency", 1, "max number of pushes to attempt in parallel")
	flag.BoolVar(&config.Silent, "silent", false, "true to run the commands and print output the terminal")
	flag.StringVar(&config.Output, "output", "", "if specified, writes benchmark results to a CSV file")
	flag.StringVar(&config.Workload, "workload", "", "The set of operations a user should issue (ex. login,push,push)")
	flag.IntVar(&config.Interval, "interval", 0, "repeat a workload at n second interval, to be used with -stop")
	flag.IntVar(&config.Stop, "stop", 0, "stop a repeating interval after n second, to be used with -interval")

	return &config
}

func (config *Config) Parse() error {
	if config.Config != "" {
		err := config.parseConfigParamFile()
		if err != nil {
			return err
		}
	}
	flag.Parse()
	return nil
}

func (config *Config) parseConfigParamFile() error {
	file, err := ioutil.ReadFile(config.Config)
	if err != nil {
		return err
	}

	err = goyaml.Unmarshal(file, &config)
	if err != nil {
		return err
	}

	return nil
}
