package redis

import (
	"errors"
	"os"

	"github.com/cloudfoundry-incubator/pat/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		connectionFromFactory Conn
		redisHost             string
		redisPort             int
		redisPassword         string
		flags                 config.Config
		args                  []string
	)

	BeforeEach(func() {
		flags = config.NewConfig()
		DescribeParameters(flags)
		args = []string{}

		connectionFromFactory = &dummyConn{}
		ConnFactory = func(host string, port int, password string) (Conn, error) {
			redisHost = host
			redisPort = port
			redisPassword = password
			return connectionFromFactory, nil
		}

		DescribeParameters(flags)
	})

	JustBeforeEach(func() {
		flags.Parse(args)
	})

	BeforeEach(func() {
		args = []string{"-redis-host", "rhost", "-redis-port", "12344", "-redis-password", "p444w"}
	})

	It("Creates a redis connection with the host, port and password", func() {
		var c Conn
		WithRedisConnection(func(conn Conn) error {
			c = conn
			return nil
		})

		Ω(c).Should(Equal(connectionFromFactory))
		Ω(redisHost).Should(Equal("rhost"))
		Ω(redisPort).Should(Equal(12344))
		Ω(redisPassword).Should(Equal("p444w"))
	})

	Context("But if the factory returns an error", func() {
		BeforeEach(func() {
			ConnFactory = func(host string, port int, password string) (Conn, error) {
				return connectionFromFactory, errors.New("Some error")
			}
		})

		It("Returns an error and does not run the passed function", func() {
			var wasRun = false
			err := WithRedisConnection(func(conn Conn) error {
				wasRun = true
				return nil
			})

			Ω(wasRun).Should(BeFalse())
			Ω(err).Should(HaveOccurred())
		})
	})

	Context("And the passed function doesn't return an error", func() {
		It("returns nil", func() {
			err := WithRedisConnection(func(conn Conn) error {
				return nil
			})

			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("And the passed function returns an error", func() {
		It("returns an error also", func() {
			err := WithRedisConnection(func(conn Conn) error {
				return errors.New("an error")
			})

			Ω(err).Should(HaveOccurred())
		})
	})

	Context("But when VCAP_SERVICES is specified", func() {
		Context("And contains a service called 'redis'", func() {
			BeforeEach(func() {
				os.Setenv("VCAP_SERVICES", `{"redis-2.2":[
						{
							"name": "redis",
							"credentials":{
								"hostname":"the-vcap-redis-host",
								"port":5004,
								"password":"vcap-redis-pass"
							}
						}]}
					`)
			})

			It("Creates a store using the credentials in VCAP_SERVICES", func() {
				var c Conn
				WithRedisConnection(func(conn Conn) error {
					c = conn
					return nil
				})

				Ω(c).Should(Equal(connectionFromFactory))
				Ω(redisHost).Should(Equal("the-vcap-redis-host"))
				Ω(redisPort).Should(Equal(5004))
				Ω(redisPassword).Should(Equal("vcap-redis-pass"))
			})
		})

		Context("And it doesn't contain a service called 'redis'", func() {
			BeforeEach(func() {
				os.Setenv("VCAP_SERVICES", `{"redis-2.2":[
						{
							"name": "NOT REDIS",
							"credentials":{
								"hostname":"the-vcap-redis-host",
								"port":5004,
								"password":"vcap-redis-pass"
							}
						}]}
					`)
			})

			It("Uses the command line values", func() {
				var c Conn
				WithRedisConnection(func(conn Conn) error {
					c = conn
					return nil
				})

				Ω(c).Should(Equal(connectionFromFactory))
				Ω(redisHost).Should(Equal("rhost"))
			})
		})
	})
})

type dummyConn struct{}

func (d dummyConn) Do(cmd string, args ...interface{}) (interface{}, error) {
	return nil, nil
}
