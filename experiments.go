package pat

import (
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/cf-acceptance-tests/helpers"
	. "github.com/vito/cmdtest/matchers"
	"time"
)

func dummy() {
	time.Sleep(1 * time.Second)
}

func push() {
	Expect(Cf("push", "patsapp", "-p", "assets/hello-world")).To(Say("App started"))
}
