package store_test

import (
	"os"

	"github.com/julz/pat/config"
	"github.com/julz/pat/laboratory"
	. "github.com/julz/pat/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		csvStoreDir   string
		csvStore      laboratory.Store
		redisHost     string
		redisPort     int
		redisPassword string
		redisStore    laboratory.Store
		flags         config.Config
		args          []string
	)

	BeforeEach(func() {
		flags = config.NewConfig()
		DescribeParameters(flags)
		args = []string{}
		csvStore = NewCsvStore("/tmp/fakecsvstore")
		redisStore = NewCsvStore("/tmp/fakeredisstore")
		CsvStoreFactory = func(dir string) laboratory.Store {
			csvStoreDir = dir
			return csvStore
		}
		RedisStoreFactory = func(host string, port int, password string) (laboratory.Store, error) {
			redisHost = host
			redisPort = port
			redisPassword = password
			return redisStore, nil
		}
	})

	JustBeforeEach(func() {
		flags.Parse(args)
	})

	Context("When useRedis is false", func() {
		BeforeEach(func() {
			args = []string{"-use-redis=false", "-csv-dir", "foo/bar/baz"}
		})

		It("Uses the csvDir paramter to configure a CSV store", func() {
			var s laboratory.Store = nil
			WithStore(func(store laboratory.Store) error {
				s = store
				return nil
			})

			Ω(s).Should(Equal(csvStore))
			Ω(csvStoreDir).Should(Equal("foo/bar/baz"))
		})
	})

	Context("When useRedis is true", func() {
		BeforeEach(func() {
			args = []string{"-use-redis", "-redis-host", "rhost", "-redis-port", "12344", "-redis-password", "p444w"}
		})

		It("Creates a redis store with the host, port and password", func() {
			var s laboratory.Store = nil
			WithStore(func(store laboratory.Store) error {
				s = store
				return nil
			})

			Ω(s).Should(Equal(redisStore))
			Ω(redisHost).Should(Equal("rhost"))
			Ω(redisPort).Should(Equal(12344))
			Ω(redisPassword).Should(Equal("p444w"))
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
					var s laboratory.Store = nil
					WithStore(func(store laboratory.Store) error {
						s = store
						return nil
					})

					Ω(s).Should(Equal(redisStore))
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
					var s laboratory.Store = nil
					WithStore(func(store laboratory.Store) error {
						s = store
						return nil
					})

					Ω(s).Should(Equal(redisStore))
					Ω(redisHost).Should(Equal("rhost"))
				})
			})
		})
	})
})
