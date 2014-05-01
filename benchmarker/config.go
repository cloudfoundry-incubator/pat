package benchmarker

import (
	"io"

	"github.com/cloudfoundry-incubator/pat/config"
	"github.com/cloudfoundry-incubator/pat/redis"
	"github.com/cloudfoundry-incubator/pat/workloads"
)

var params = struct {
	startMasterAndSlave bool
}{}

func DescribeParameters(config config.Config) {
	config.BoolVar(&params.startMasterAndSlave, "use-redis-worker", false, "Runs in master mode, sending work to perform to a redis queue")
	WorkloadListFactory().DescribeParameters(config)
}

func WithConfiguredWorkerAndSlaves(fn func(worker Worker) error) error {
	if params.startMasterAndSlave {
		return WithRedisConnection(func(conn redis.Conn) error {
			slave := SlaveFactory(conn, configure(LocalWorkerFactory()))
			defer slave.Close()
			return fn(configure(RedisWorkerFactory(conn)))
		})
	}

	return fn(configure(LocalWorkerFactory()))
}

func configure(worker Worker) Worker {
	workloadList := WorkloadListFactory()
	workloadList.DescribeWorkloads(worker)
	return worker
}

type WorkloadDescriber interface {
	DescribeWorkloads(worker workloads.WorkloadAdder)
	DescribeParameters(config config.Config)
}

var WithRedisConnection = func(fn func(conn redis.Conn) error) error {
	return redis.WithRedisConnection(fn)
}

var LocalWorkerFactory = func() *LocalWorker {
	return NewLocalWorker()
}

var RedisWorkerFactory = func(conn redis.Conn) Worker {
	return NewRedisWorker(conn)
}

var SlaveFactory = func(conn redis.Conn, delegate Worker) io.Closer {
	return StartSlave(conn, delegate)
}

var WorkloadListFactory = func() WorkloadDescriber {
	return workloads.DefaultWorkloadList()
}
