package store_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/pat/experiment"
	"github.com/cloudfoundry-incubator/pat/redis"
	. "github.com/cloudfoundry-incubator/pat/store"
	redisHelpers "github.com/cloudfoundry-incubator/pat/test_helpers/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type store interface {
	LoadAll() ([]experiment.Experiment, error)
	Writer(name string) func(samples <-chan *experiment.Sample)
}

var _ = Describe("Redis Store", func() {
	var (
		store store
	)

	BeforeEach(func() {
		redisHelpers.StartRedis("redis.conf")
	})

	AfterEach(func() {
		redisHelpers.StopRedis()
	})

	Describe("Saving and Loading", func() {
		BeforeEach(func() {
			conn, err := redis.Connect("", 63798, "p4ssw0rd")
			Ω(err).ShouldNot(HaveOccurred())
			store, err = NewRedisStore(conn)
			Ω(err).ShouldNot(HaveOccurred())

			writer := store.Writer("experiment-1")
			write(writer, []*experiment.Sample{
				&experiment.Sample{nil, 1, 2, 3, 4, 5, 6, nil, 7, 9, 8, experiment.ResultSample},
				&experiment.Sample{nil, 9, 8, 7, 6, 5, 4, errors.New("foo"), 3, 1, 2, experiment.ResultSample},
			})

			writer = store.Writer("experiment-2")
			write(writer, []*experiment.Sample{
				&experiment.Sample{nil, 2, 2, 3, 4, 5, 6, nil, 7, 9, 8, experiment.ResultSample},
			})

			writer = store.Writer("experiment-3")
			write(writer, []*experiment.Sample{
				&experiment.Sample{nil, 1, 3, 3, 4, 5, 6, nil, 7, 9, 8, experiment.ResultSample},
				&experiment.Sample{nil, 2, 3, 3, 4, 5, 6, nil, 7, 9, 8, experiment.ResultSample},
				&experiment.Sample{nil, 9, 8, 7, 6, 5, 4, errors.New("foo"), 3, 1, 2, experiment.ResultSample},
			})

			writer = store.Writer("experiment-with-no-data")
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
