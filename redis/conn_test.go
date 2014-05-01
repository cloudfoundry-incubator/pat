package redis_test

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
	. "github.com/cloudfoundry-incubator/pat/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Conn", func() {
	BeforeEach(func() {
		StartRedis("redis.conf")
	})

	AfterEach(func() {
		StopRedis()
	})

	Describe("Connecting", func() {
		Context("When the host is wrong", func() {
			It("Returns an error", func() {
				_, err := Connect("fishfinger", 63798, "p4ssw0rd")
				Ω(err).Should(HaveOccurred())
			})
		})

		Context("When the port is wrong", func() {
			It("Returns an error", func() {
				_, err := Connect("localhost", 63799, "p4ssw0rd")
				Ω(err).Should(HaveOccurred())
			})
		})

		Context("When the host and port are correct", func() {
			Context("But the password is wrong", func() {
				It("Returns an error", func() {
					_, err := Connect("localhost", 63799, "WRONG")
					Ω(err).Should(HaveOccurred())
				})
			})

			Context("And the password is correct", func() {
				It("works", func() {
					_, err := Connect("localhost", 63798, "p4ssw0rd")
					Ω(err).ShouldNot(HaveOccurred())
				})
			})

			Context("When the server has no password", func() {
				It("works", func() {
					StopRedis()
					StartRedis("redis.nopass.conf")
					_, err := Connect("localhost", 63798, "")
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
		})
	})
})

func StartRedis(config string) {
	_, filename, _, _ := runtime.Caller(0)
	dir, _ := filepath.Abs(filepath.Dir(filename))
	StopRedis()
	exec.Command("redis-server", dir+"/"+config).Run()
	time.Sleep(450 * time.Millisecond) // yuck(jz)
}

func StopRedis() {
	exec.Command("redis-cli", "-p", "63798", "shutdown").Run()
	exec.Command("redis-cli", "-p", "63798", "-a", "p4ssw0rd", "shutdown").Run()
}
