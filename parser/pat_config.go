package parser

import (
	"io/ioutil"
	"launchpad.net/goyaml"
)

type PATs struct {
	Cli_commands struct {
		Server      bool
		Pushes      int
		Concurrency int
		Silent      bool
		Output      string
		Interval    int
		Stop        int
	}
}

func NewPATsConfiguration(fName string) (*PATs, error) {
	var pat = PATs{}

	file, err := ioutil.ReadFile(fName)
	if err != nil {
		return nil, err
	}

	err = goyaml.Unmarshal(file, &pat)
	if err != nil {
		return nil, err
	}

	return &pat, nil
}
