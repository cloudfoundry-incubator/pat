package store_test

import (
	"errors"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/julz/pat/experiment"
	. "github.com/julz/pat/store"
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
		err   error
	)

	BeforeEach(func() {
		_, filename, _, _ := runtime.Caller(0)
		dir, _ := filepath.Abs(filepath.Dir(filename))
		exec.Command("redis-cli", "-p", "63798", "shutdown").Run()
		exec.Command("redis-server", dir+"/redis.conf").Run()
		time.Sleep(500 * time.Millisecond) // yuck(jz)
	})

	AfterEach(func() {
		exec.Command("redis-cli", "-p", "63798", "-a", "p4ssw0rd", "shutdown").Run()
	})

	Describe("Connecting", func() {
		Context("When the host is wrong", func() {
			It("Returns an error", func() {
				_, err := NewRedisStore("fishfinger", 63798, "p4ssw0rd")
				Ω(err).Should(HaveOccurred())
			})
		})

		Context("When the port is wrong", func() {
			It("Returns an error", func() {
				_, err := NewRedisStore("localhost", 63799, "p4ssw0rd")
				Ω(err).Should(HaveOccurred())
			})
		})

		Context("When the host and port are correct", func() {
			Context("But the password is wrong", func() {
				It("Returns an error", func() {
					_, err := NewRedisStore("localhost", 63799, "WRONG")
					Ω(err).Should(HaveOccurred())
				})
			})

			Context("And the password is correct", func() {
				It("works", func() {
					_, err := NewRedisStore("localhost", 63798, "p4ssw0rd")
					Ω(err).ShouldNot(HaveOccurred())
				})
			})
		})
	})

	Describe("Saving and Loading", func() {
		BeforeEach(func() {
			store, err = NewRedisStore("", 63798, "p4ssw0rd")
			Ω(err).ShouldNot(HaveOccurred())

			writer := store.Writer("experiment-1")
			write(writer, []*experiment.Sample{
				&experiment.Sample{nil, 1, 2, 3, 4, 5, 6, nil, 7, 8, experiment.ResultSample},
				&experiment.Sample{nil, 9, 8, 7, 6, 5, 4, errors.New("foo"), 3, 2, experiment.ResultSample},
			})

			writer = store.Writer("experiment-2")
			write(writer, []*experiment.Sample{
				&experiment.Sample{nil, 2, 2, 3, 4, 5, 6, nil, 7, 8, experiment.ResultSample},
			})

			writer = store.Writer("experiment-3")
			write(writer, []*experiment.Sample{
				&experiment.Sample{nil, 1, 3, 3, 4, 5, 6, nil, 7, 8, experiment.ResultSample},
				&experiment.Sample{nil, 2, 3, 3, 4, 5, 6, nil, 7, 8, experiment.ResultSample},
				&experiment.Sample{nil, 9, 8, 7, 6, 5, 4, errors.New("foo"), 3, 2, experiment.ResultSample},
			})
		})

		It("Round trips experiment list", func() {
			experiments, err := store.LoadAll()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(experiments).Should(HaveLen(3))
		})

		It("Round trips experiment guids", func() {
			experiments, err := store.LoadAll()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(experiments[0].GetGuid()).Should(Equal("experiment-1"))
			Ω(experiments[1].GetGuid()).Should(Equal("experiment-2"))
			Ω(experiments[2].GetGuid()).Should(Equal("experiment-3"))
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
	})
})
