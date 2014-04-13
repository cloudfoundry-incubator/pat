package logs

import (
	"os"
	"strings"

	"github.com/cloudfoundry-community/pat/config"
	"github.com/cloudfoundry/gosteno"
)

var params = struct {
	path  string
	level string
}{}

func InitCommandLineFlags(flags config.Config) {
	flags.StringVar(&params.path, "logging:file", "", "A file to log to, or empty to log to STDOUT")
	flags.StringVar(&params.level, "logging:level", "INFO", "The level to log at, one of all, debug2, debug1, debug, info, warn, error, fatal, off")
}

var initialized bool

func NewLogger(name string) *gosteno.Logger {
	if !initialized {
		initLogging()
		initialized = true
	}

	return gosteno.NewLogger(name)
}

func initLogging() {
	level, err := gosteno.GetLogLevel(strings.ToLower(params.level))
	if err != nil {
		panic(err)
	}

	sinks := []gosteno.Sink{}
	if params.path != "" {
		sinks = append(sinks, NewFileSink(params.path))
	} else {
		sinks = append(sinks, NewIOSink(os.Stdout))
	}

	c := &gosteno.Config{
		Sinks:     sinks,
		Level:     level,
		EnableLOC: true,
	}

	InitGoSteno(c)
}

var InitGoSteno = func(c *gosteno.Config) {
	gosteno.Init(c)
}

var NewFileSink = func(path string) gosteno.Sink {
	return gosteno.NewFileSink(path)
}

var NewIOSink = func(file *os.File) gosteno.Sink {
	return gosteno.NewIOSink(file)
}
