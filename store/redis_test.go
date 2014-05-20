package store_test

import (
	"encoding/json"
	"time"
	"errors"

	"github.com/cloudfoundry-incubator/pat/experiment"
	"github.com/cloudfoundry-incubator/pat/redis"
	"github.com/cloudfoundry-incubator/pat/test_helpers"
	. "github.com/cloudfoundry-incubator/pat/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type store interface {
	LoadAll() ([]experiment.Experiment, error)
	Writer(ex experiment.ExperimentConfiguration) func(samples <-chan *experiment.Sample)
}

var _ = Describe("Redis Store", func() {
	var (
		store                   store
		experimentConfiguration experiment.ExperimentConfiguration
	)

	BeforeEach(func() {
		test_helpers.StartRedis("redis.conf")
	})

	AfterEach(func() {
		test_helpers.StopRedis()
	})

	Describe("Experiments", func() {
		Context("Saving and Loading", func() {
			BeforeEach(func() {
				conn, err := redis.Connect("", 63798, "p4ssw0rd")
				Ω(err).ShouldNot(HaveOccurred())
				store, err = NewRedisStore(conn)
				Ω(err).ShouldNot(HaveOccurred())

				writer := store.Writer(experiment.ExperimentConfiguration{Guid: "experiment-1"})
				write(writer, []*experiment.Sample{
					&experiment.Sample{nil, 1, 2, 3, 4, 5, 6, nil, 7, 9, 8, experiment.ResultSample},
					&experiment.Sample{nil, 9, 8, 7, 6, 5, 4, errors.New("foo"), 3, 1, 2, experiment.ResultSample},
				})

				writer = store.Writer(experiment.ExperimentConfiguration{Guid: "experiment-2"})
				write(writer, []*experiment.Sample{
					&experiment.Sample{nil, 2, 2, 3, 4, 5, 6, nil, 7, 9, 8, experiment.ResultSample},
				})

				writer = store.Writer(experiment.ExperimentConfiguration{Guid: "experiment-3"})
				write(writer, []*experiment.Sample{
					&experiment.Sample{nil, 1, 3, 3, 4, 5, 6, nil, 7, 9, 8, experiment.ResultSample},
					&experiment.Sample{nil, 2, 3, 3, 4, 5, 6, nil, 7, 9, 8, experiment.ResultSample},
					&experiment.Sample{nil, 9, 8, 7, 6, 5, 4, errors.New("foo"), 3, 1, 2, experiment.ResultSample},
				})

				writer = store.Writer(experiment.ExperimentConfiguration{Guid: "experiment-with-no-data"})
			})

			It("Round trips experiment list", func() {
				experiments, err := store.LoadAll()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(experiments).Should(HaveLen(4))
			})

			It("Round trips experiment guids", func() {
				experiments, err := store.LoadAll()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(experiments[0].GetGuid()).Should(Equal("experiment-1"))
				Ω(experiments[1].GetGuid()).Should(Equal("experiment-2"))
				Ω(experiments[2].GetGuid()).Should(Equal("experiment-3"))
				Ω(experiments[3].GetGuid()).Should(Equal("experiment-with-no-data"))
			})

			It("Round trips samples", func() {
				experiments, err := store.LoadAll()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(data(experiments[0].GetData())).Should(HaveLen(2))
				Ω(data(experiments[1].GetData())).Should(HaveLen(1))
				Ω(data(experiments[2].GetData())).Should(HaveLen(3))
			})

			It("Round trips sample data", func() {
				experiments, err := store.LoadAll()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(data(experiments[0].GetData())[1].Total).Should(Equal(int64(7)))
				Ω(data(experiments[1].GetData())[0].TotalErrors).Should(Equal(4))
				Ω(data(experiments[2].GetData())[2].TotalWorkers).Should(Equal(5))
			})

			It("Returns empty array if data not found (redis cannot distinguish empty from not-created lists)", func() {
				experiments, err := store.LoadAll()
				Ω(err).ShouldNot(HaveOccurred())
				Ω(experiments[3].GetGuid()).Should(Equal("experiment-with-no-data"))
				Ω(data(experiments[3].GetData())).ShouldNot(BeNil())
				Ω(data(experiments[3].GetData())).Should(HaveLen(0))
			})
		})
	})

	Describe("Meta data", func() {
		Context("Saving", func() {
			const (
				iterations          = 2
				concurrencyStepTime = time.Duration(5)
				interval            = 10
				stop                = 100
				workload            = "gcf:push"
				note                = "note description"
			)

			var (
				meta	MetaData
				conn	redis.Conn
				err	error
				concurrency = []int{1, 2, 3}
			)

			BeforeEach(func() {
				conn, err = redis.Connect("", 63798, "p4ssw0rd")
				Ω(err).ShouldNot(HaveOccurred())
				store, err = NewRedisStore(conn)
				Ω(err).ShouldNot(HaveOccurred())

				experimentConfiguration = experiment.NewExperimentConfiguration(
					iterations, concurrency, concurrencyStepTime,
					interval, stop, nil, workload, note)

				store.Writer(experimentConfiguration)
			})

			It("Should have created the meta_data key", func() {
				data, err := conn.Do("EXISTS", "meta_data")
				Ω(err).Should(BeNil())
				Ω(data).Should(Equal(int64(1)))
			})

			It("Should save the guid in the meta data", func() {
				data, err := redis.Strings(conn.Do("LRANGE", "meta_data", 0, 0))
				err = json.Unmarshal([]byte(data[0]), &meta)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(meta.Guid).Should(Equal(experimentConfiguration.Guid))
			})

			It("Should save the concurrency meta data", func() {
				data, _ := redis.Strings(conn.Do("LRANGE", "meta_data", 0, 0))
				err := json.Unmarshal([]byte(data[0]), &meta)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(meta.Concurrency).Should(Equal("1..2..3"))
			})

			It("Should save the iteration meta data", func() {
				data, _ := redis.Strings(conn.Do("LRANGE", "meta_data", 0, 0))
				err := json.Unmarshal([]byte(data[0]), &meta)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(meta.Iterations).Should(Equal(iterations))
			})

			It("Should save the stop meta data", func() {
				data, _ := redis.Strings(conn.Do("LRANGE", "meta_data", 0, 0))
				err := json.Unmarshal([]byte(data[0]), &meta)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(meta.Stop).Should(Equal(stop))
			})

			It("Should save the interval meta data", func() {
				data, _ := redis.Strings(conn.Do("LRANGE", "meta_data", 0, 0))
				err := json.Unmarshal([]byte(data[0]), &meta)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(meta.Interval).Should(Equal(interval))
			})

			It("Should save the workload meta data", func() {
				data, _ := redis.Strings(conn.Do("LRANGE", "meta_data", 0, 0))
				err := json.Unmarshal([]byte(data[0]), &meta)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(meta.Workload).Should(Equal(workload))
			})

			It("Should save the note meta data", func() {
				data, _ := redis.Strings(conn.Do("LRANGE", "meta_data", 0, 0))
				err := json.Unmarshal([]byte(data[0]), &meta)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(meta.Note).Should(Equal(note))
			})
		})
	})
})
