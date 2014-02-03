package experiments

import (
	"errors"
	"math/rand"
	"time"

	"github.com/nu7hatch/gouuid"
	. "github.com/pivotal-cf-experimental/cf-acceptance-tests/helpers"
)

//Todo(simon) Remove, for dev testing only
func random(min, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	r := min + rand.Intn(max-min)
	return r
}

func Dummy() error {
	time.Sleep(time.Duration(random(1, 5)) * time.Second)
	return nil
}

func DummyWithErrors() error {
	Dummy()
	if random(0, 10) > 8 {
		return errors.New("Random (dummy) error")
	}
	return nil
}

func Push() error {
	guid, _ := uuid.NewV4()
	err := Cf("push", "pats-"+guid.String(), "patsapp", "-m", "64M", "-p", "assets/hello-world").ExpectOutput("App started")
	return err
}
