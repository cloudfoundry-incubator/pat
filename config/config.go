package config

import (
	"flag"
	"io/ioutil"
	"os"

	"launchpad.net/goyaml"
)

type Config interface {
	StringVar(target *string, name string, defaultValue string, description string)
	IntVar(target *int, name string, defaultValue int, description string)
	BoolVar(target *bool, name string, defaultValue bool, description string)
	Parse(args []string) error
}

type f struct {
	flagSet *flag.FlagSet
}

func NewConfig() *f {
	return &f{flag.NewFlagSet(os.Args[0], flag.ExitOnError)}
}

var ConfigAndFlags = NewConfig()

func (f *f) StringVar(target *string, name string, defaultValue string, description string) {
	f.flagSet.StringVar(target, name, defaultValue, description)
}

func (f *f) IntVar(target *int, name string, defaultValue int, description string) {
	f.flagSet.IntVar(target, name, defaultValue, description)
}

func (f *f) BoolVar(target *bool, name string, defaultValue bool, description string) {
	f.flagSet.BoolVar(target, name, defaultValue, description)
}

func (f *f) Parse(args []string) error {
	config := f.flagSet.String("config", "", "YML file containing configuration parameters")

	f.flagSet.Parse(args)
	if len(*config) > 0 {
		if err := f.ParseConfig(*config); err != nil {
			panic("Failed Parsing Config File")
		}
	}

	return nil
}

func (f *f) ParseConfig(path string) error {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	yml := make(map[string]string)
	err = goyaml.Unmarshal(file, &yml)
	if err != nil {
		return err
	}

	f.flagSet.Visit(func(flag *flag.Flag) {
		delete(yml, flag.Name)
	})

	for k, v := range yml {
		flag := f.flagSet.Lookup(k)
		flag.Value.Set(v)
	}

	return nil
}
