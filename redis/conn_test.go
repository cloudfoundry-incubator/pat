package redis_test

import (
	redisHelpers "github.com/cloudfoundry-incubator/pat/test_helpers/redis"

	. "github.com/cloudfoundry-incubator/pat/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Conn", func() {
	BeforeEach(func() {
		redisHelpers.StartRedis("redis.conf")
	})

	AfterEach(func() {
		redisHelpers.StopRedis()
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
					redisHelpers.StopRedis()
					redisHelpers.StartRedis("redis.nopass.conf")
					_, err := Connect("localhost", 63798, "")
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
		})
	})
})
