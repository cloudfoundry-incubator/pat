package benchmarker

import (
	"errors"
	"io"

	"github.com/cloudfoundry-community/pat/config"
	"github.com/cloudfoundry-community/pat/redis"
	"github.com/cloudfoundry-community/pat/workloads"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		flags                 config.Config
		args                  []string
		localWorker           *LocalWorker
		connectionFromFactory redis.Conn
		redisWorkerConn       redis.Conn
		redisWorker           Worker
		slaveFromFactory      *dummySlave
		slaveStarted          bool
	)

	BeforeEach(func() {
		flags = config.NewConfig()
		DescribeParameters(flags)
		args = []string{}

		localWorker = NewLocalWorker()
		LocalWorkerFactory = func() *LocalWorker {
			return localWorker
		}

		redisWorker = NewLocalWorker()
		redisWorker.AddWorkloadStep(workloads.WorkloadStep{"redis", nil, "b"})
		redisWorkerConn = &dummyConn{"redisConn"}
		RedisWorkerFactory = func(conn redis.Conn) Worker {
			redisWorkerConn = conn
			return redisWorker
		}

		connectionFromFactory = &dummyConn{"cff"}
		WithRedisConnection = func(fn func(conn redis.Conn) error) error {
			return fn(connectionFromFactory)
		}

		slaveStarted = false
		SlaveFactory = func(conn redis.Conn, worker Worker) io.Closer {
			slaveFromFactory = &dummySlave{conn, worker, false}
			slaveStarted = true
			return slaveFromFactory
		}

		WorkloadListFactory = func() WorkloadDescriber {
			return &dummyDescriberWithThreeWorkloads{}
		}
	})

	JustBeforeEach(func() {
		flags.Parse(args)
	})

	Context("When -use-redis-worker is not set", func() {
		It("Calls with a local worker", func() {
			var worker Worker
			WithConfiguredWorkerAndSlaves(func(w Worker) error {
				worker = w
				return nil
			})

			Ω(worker).Should(Equal(localWorker))
			Ω(worker).ShouldNot(Equal(redisWorker))
		})

		It("doesn't start a slave", func() {
			WithConfiguredWorkerAndSlaves(func(w Worker) error {
				return nil
			})

			Ω(slaveStarted).Should(BeFalse())
		})

		It("Configures the worker with default workloads", func() {
			var worker Worker
			WithConfiguredWorkerAndSlaves(func(w Worker) error {
				worker = w
				return nil
			})

			Ω(worker.(*LocalWorker).Experiments).Should(HaveLen(3))
		})

		Context("And if it returns an error", func() {
			It("should return it", func() {
				err := WithConfiguredWorkerAndSlaves(func(w Worker) error {
					return errors.New("Some error")
				})

				Ω(err).Should(HaveOccurred())
			})
		})
	})

	Context("When -use-redis-worker is set", func() {
		BeforeEach(func() {
			args = []string{"-use-redis-worker", "true"}
		})

		It("Configures and calls with a redis worker", func() {
			var worker Worker
			WithConfiguredWorkerAndSlaves(func(w Worker) error {
				worker = w
				return nil
			})

			Ω(redisWorkerConn).Should(Equal(connectionFromFactory))
			Ω(worker).Should(Equal(redisWorker))
			Ω(worker).ShouldNot(Equal(localWorker))
		})

		It("Configures the worker stub with default workloads, so that -list/-validate-workloads works properly", func() {
			var worker Worker
			WithConfiguredWorkerAndSlaves(func(w Worker) error {
				worker = w
				return nil
			})

			Ω(worker.(*LocalWorker).Experiments).Should(HaveLen(4))
		})

		Context("And if it returns an error", func() {
			It("should return it", func() {
				err := WithConfiguredWorkerAndSlaves(func(w Worker) error {
					return errors.New("Some error")
				})

				Ω(err).Should(HaveOccurred())
			})
		})

		It("starts a slave", func() {
			WithConfiguredWorkerAndSlaves(func(w Worker) error {
				return nil
			})

			Ω(slaveStarted).Should(BeTrue())
			Ω(slaveFromFactory.conn).Should(Equal(connectionFromFactory))
			Ω(slaveFromFactory.worker).Should(Equal(localWorker))
		})

		It("closes the slave after the function returns", func() {
			Ω(slaveFromFactory.wasClosed).Should(BeTrue())
		})
	})
})

type dummyConn struct{ name string }

func (dummyConn) Do(cmd string, args ...interface{}) (interface{}, error) {
	return nil, nil
}

type dummySlave struct {
	conn      redis.Conn
	worker    Worker
	wasClosed bool
}

func (d *dummySlave) Close() error {
	d.wasClosed = true
	return nil
}

type dummyDescriberWithThreeWorkloads struct{}

func (dummyDescriberWithThreeWorkloads) DescribeWorkloads(worker workloads.WorkloadAdder) {
	worker.AddWorkloadStep(workloads.Step("a", nil, "desc"))
	worker.AddWorkloadStep(workloads.Step("b", nil, "desc"))
	worker.AddWorkloadStep(workloads.Step("c", nil, "desc"))
}
