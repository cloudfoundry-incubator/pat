package store_test

import (
	"github.com/cloudfoundry-incubator/pat/config"
	"github.com/cloudfoundry-incubator/pat/ginkgo/redis_helpers"
	"github.com/cloudfoundry-incubator/pat/laboratory"
	"github.com/cloudfoundry-incubator/pat/redis"
	. "github.com/cloudfoundry-incubator/pat/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		args  []string
		flags config.Config
	)

	BeforeEach(func() {
		flags = config.NewConfig()
		DescribeParameters(flags)
	})

	JustBeforeEach(func() {
		flags.Parse(args)
	})

	Describe("#WithStore", func() {
		var (
			csvStoreDir     string
			csvStore        laboratory.Store
			connFromFactory redis.Conn
			redisConn       redis.Conn
			redisStore      laboratory.Store
		)

		BeforeEach(func() {
			csvStore = NewCsvStore("/tmp/fakecsvstore", nil)
			redisStore = NewCsvStore("/tmp/fakeredisstore", nil)
			CsvStoreFactory = func(dir string) laboratory.Store {
				csvStoreDir = dir
				return csvStore
			}

			connFromFactory = &dummyConn{}
			WithRedisConnection = func(fn func(conn redis.Conn) error) error {
				return fn(connFromFactory)
			}

			RedisStoreFactory = func(conn redis.Conn) (laboratory.Store, error) {
				redisConn = conn
				return redisStore, nil
			}
		})

		Context("When use-redis-store is false", func() {
			BeforeEach(func() {
				args = []string{"-use-redis-store=false", "-csv-dir", "foo/bar/baz"}
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
				args = []string{"-use-redis-store"}
			})

			It("Creates a Redis store using the connection from redis.WithRedisConnection(", func() {
				var s laboratory.Store = nil
				WithStore(func(store laboratory.Store) error {
					s = store
					return nil
				})

				Ω(s).Should(Equal(redisStore))
				Ω(redisConn).Should(Equal(connFromFactory))
			})
		})
	})

	Describe("#MetaStoreFactory", func() {
		const (
			dir = "tmp"
		)

		var (
			meta *MetaStore
		)

		It("Returns a pointer to a MetaStore struct", func() {
			meta, _ = MetaStoreFactory(dir)
			Ω(meta).Should(BeAssignableToTypeOf(&MetaStore{}))
		})

		Context("When use-redis-store is false", func() {
			BeforeEach(func() {
				args = []string{"-use-redis-store=false"}
			})

			It("sets UseRedis to false", func() {
				meta, _ = MetaStoreFactory(dir)
				Ω(meta.UseRedis).Should(Equal(false))
			})

			It("does not return a redis connection", func() {
				meta, _ = MetaStoreFactory(dir)
				Ω(meta.Conn).Should(BeNil())
			})
		})

		Context("When use-redis-store is true", func() {
			BeforeEach(func() {
				args = []string{"-use-redis-store=true", "-redis-host", "localhost", "-redis-port", "6379"}

				redis_helpers.StartRedis("redis_local.conf")
				err := redis_helpers.CheckRedisRunning()
				Ω(err).Should(BeNil())
			})

			AfterEach(func() {
				redis_helpers.StopLocalRedis()
			})

			It("sets UseRedis to true", func() {
				meta, _ = MetaStoreFactory(dir)
				Ω(meta.UseRedis).Should(Equal(true))
			})

			It("sets Conn to a redis connection", func() { //(Dan) hard to test for
				_, err := MetaStoreFactory(dir)
				Ω(err).Should(BeNil())
			})
		})
	})
})

type dummyConn struct{}

func (dummyConn) Do(cmd string, args ...interface{}) (interface{}, error) {
	return nil, nil
}
