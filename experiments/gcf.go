package experiments

import (
	"github.com/nu7hatch/gouuid"
	. "github.com/pivotal-cf-experimental/cf-acceptance-tests/helpers"
	"time"
)

func Dummy() error {
	time.Sleep(3 * time.Second)
	return nil
}

func Push() error {
	guid, _ := uuid.NewV4()
	err := Cf("push", "pats-"+guid.String(), "patsapp", "-m", "64M", "-p", "assets/hello-world").ExpectOutput("App started")
	return err
}
