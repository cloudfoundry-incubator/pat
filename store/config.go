package store

import (
	"github.com/cloudfoundry-community/pat/config"
	"github.com/cloudfoundry-community/pat/laboratory"
	"github.com/cloudfoundry-community/pat/redis"
	"github.com/cloudfoundry-community/pat/workloads"
)

var params = struct {
	csvDir   string
	useRedis bool
}{}

func DescribeParameters(config config.Config) {
	config.StringVar(&params.csvDir, "csv-dir", "output/csvs", "Directory to Store CSVs")
	config.BoolVar(&params.useRedis, "use-redis-store", false, "True if redis should be used (requires the -redis-host, -redis-port and -redis-password arguments)")
	redis.DescribeParameters(config)
}

func WithStore(fn func(store laboratory.Store) error) error {
	if params.useRedis {
		return WithRedisConnection(func(conn redis.Conn) error {
			store, err := RedisStoreFactory(conn)
			if err != nil {
				return err
			}

			return fn(store)
		})
	} else {
		return fn(CsvStoreFactory(params.csvDir))
	}
}

var WithRedisConnection = func(fn func(conn redis.Conn) error) error {
	return redis.WithRedisConnection(fn)
}

var RedisStoreFactory = func(conn redis.Conn) (laboratory.Store, error) {
	return NewRedisStore(conn)
}

var CsvStoreFactory = func(dir string) laboratory.Store {
	return NewCsvStore(dir, workloads.DefaultWorkloadList())
}
