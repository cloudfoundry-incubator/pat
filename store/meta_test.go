package store_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/pat/config"
	"github.com/cloudfoundry-incubator/pat/redis"
	. "github.com/cloudfoundry-incubator/pat/store"
	redisHelpers "github.com/cloudfoundry-incubator/pat/test_helpers/redis"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Meta", func() {
	Describe("#Write", func() {
		const (
			directory   = "tmp/meta"
			name        = "guid-1"
			concurrency = 1
			iterations  = 2
			interval    = 10
			stop        = 100
			workload    = "gcf:push"
			note        = "note description"
		)

		var (
			args  []string
			flags config.Config
		)

		BeforeEach(func() {
			flags = config.NewConfig()
			DescribeParameters(flags)
		})

		Context("locally", func() {
			var (
				output     string
				meta_store *MetaStore
				err        error
			)

			BeforeEach(func() {
				args = []string{"-use-redis-store=false"}
				flags.Parse(args)

				meta_store, err = MetaStoreFactory(directory)
				Ω(err).ShouldNot(HaveOccurred())
				err = meta_store.Write(name, concurrency, iterations, interval, stop, workload, note)
				Ω(err).ShouldNot(HaveOccurred())

				in, err := ioutil.ReadFile(path.Join(directory, name+".meta"))
				Ω(err).ShouldNot(HaveOccurred())
				output = string(in)
			})

			AfterEach(func() {
				os.RemoveAll("tmp")
			})

			It("saves an experiment's meta data as the first row in the csv file", func() {
				Ω(strings.Split(output, "\n")[0]).Should(ContainSubstring("start time,concurrency,iterations,stop,interval,workload,note"))
			})

			It("saves the time of the experiment as the first item in the meta data", func() {
				data := strings.Split(output, "\n")[1]
				t, err := time.Parse(time.RFC850, strings.Split(data, "\"")[1])
				Ω(err).ShouldNot(HaveOccurred())
				Ω(t).Should(BeAssignableToTypeOf(time.Time{}))
			})

			It("saves the concurrency meta data after time", func() {
				data := strings.Split(output, "\n")[1]
				Ω(strings.Split(data, ",")[2]).Should(Equal(strconv.Itoa(concurrency)))
			})

			It("saves the iteration meta data after concurrency", func() {
				data := strings.Split(output, "\n")[1]
				Ω(strings.Split(data, ",")[3]).Should(Equal(strconv.Itoa(iterations)))
			})

			It("saves the stop meta data after iterations", func() {
				data := strings.Split(output, "\n")[1]
				Ω(strings.Split(data, ",")[4]).Should(Equal(strconv.Itoa(stop)))
			})

			It("saves the interval meta data after stop", func() {
				data := strings.Split(output, "\n")[1]
				Ω(strings.Split(data, ",")[5]).Should(Equal(strconv.Itoa(interval)))
			})

			It("saves the workload meta data after interval", func() {
				data := strings.Split(output, "\n")[1]
				Ω(strings.Split(data, ",")[6]).Should(Equal(workload))
			})

			It("saves the note meta data after note", func() {
				data := strings.Split(output, "\n")[1]
				Ω(strings.Split(data, ",")[7]).Should(Equal(note))
			})
		})

		Context("using redis", func() {
			var (
				meta_store *MetaStore
				meta_data  MetaData
			)

			BeforeEach(func() {
				redisHelpers.StartRedis("redis_local.conf")
				err := redisHelpers.CheckRedisRunning()
				Ω(err).Should(BeNil())

				args := []string{"-use-redis-store=true", "-redis-host", "localhost", "-redis-port", "6379"}
				flags.Parse(args)

				meta_store, _ = MetaStoreFactory(directory)
				err = meta_store.Write(name, concurrency, iterations, interval, stop, workload, note)
				Ω(err).ShouldNot(HaveOccurred())
			})

			AfterEach(func() {
				redisHelpers.StopLocalRedis()
			})

			It("Should have written the meta file for given experiment name", func() {
				data, err := redis.Strings(meta_store.Conn.Do("KEYS", "*"))
				Ω(err).Should(BeNil())
				Ω(data[0]).Should(Equal("experiment.guid-1.meta"))
			})

			It("Should save the start time of the experiment in the meta data", func() {
				data, _ := redis.Strings(meta_store.Conn.Do("LRANGE", "experiment.guid-1.meta", 0, 0))
				err := json.Unmarshal([]byte(data[0]), &meta_data)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(time.Parse(time.RFC850, meta_data.StartTime)).Should(BeAssignableToTypeOf(time.Time{}))
			})

			It("Should save the concurrency meta data", func() {
				data, _ := redis.Strings(meta_store.Conn.Do("LRANGE", "experiment.guid-1.meta", 0, 0))
				err := json.Unmarshal([]byte(data[0]), &meta_data)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(meta_data.Concurrency).Should(Equal(concurrency))
			})

			It("Should save the iteration meta data", func() {
				data, _ := redis.Strings(meta_store.Conn.Do("LRANGE", "experiment.guid-1.meta", 0, 0))
				err := json.Unmarshal([]byte(data[0]), &meta_data)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(meta_data.Iterations).Should(Equal(iterations))
			})

			It("Should save the stop meta data", func() {
				data, _ := redis.Strings(meta_store.Conn.Do("LRANGE", "experiment.guid-1.meta", 0, 0))
				err := json.Unmarshal([]byte(data[0]), &meta_data)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(meta_data.Stop).Should(Equal(stop))
			})

			It("Should save the interval meta data", func() {
				data, _ := redis.Strings(meta_store.Conn.Do("LRANGE", "experiment.guid-1.meta", 0, 0))
				err := json.Unmarshal([]byte(data[0]), &meta_data)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(meta_data.Interval).Should(Equal(interval))
			})

			It("Should save the workload meta data", func() {
				data, _ := redis.Strings(meta_store.Conn.Do("LRANGE", "experiment.guid-1.meta", 0, 0))
				err := json.Unmarshal([]byte(data[0]), &meta_data)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(meta_data.Workload).Should(Equal(workload))
			})

			It("Should save the note meta data", func() {
				data, _ := redis.Strings(meta_store.Conn.Do("LRANGE", "experiment.guid-1.meta", 0, 0))
				err := json.Unmarshal([]byte(data[0]), &meta_data)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(meta_data.Note).Should(Equal(note))
			})
		})
	})
})
