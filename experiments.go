package pat

import (
	"github.com/nu7hatch/gouuid"
	. "github.com/pivotal-cf-experimental/cf-acceptance-tests/helpers"
	"time"
)

func dummy() error {
	time.Sleep(1 * time.Second)
	return nil
}

func push() error {
	guid, _ := uuid.NewV4()
	err := Cf("push", "pats-"+guid.String(), "patsapp", "-p", "assets/hello-world").ExpectOutput("App started")
	return err
}
