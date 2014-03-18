package store

import (
	"github.com/julz/pat/config"
	"github.com/julz/pat/laboratory"
)

var params = struct {
	csvDir        string
	useRedis      bool
	redisHost     string
	redisPort     int
	redisPassword string
}{}

func DescribeParameters(config config.Config) {
	config.StringVar(&params.csvDir, "csv-dir", "output/csvs", "Directory to Store CSVs")
	config.BoolVar(&params.useRedis, "use-redis", false, "True if redis should be used (requires the -redis-host, -redis-port and -redis-password arguments)")
	config.StringVar(&params.redisHost, "redis-host", "localhost", "Redis hostname")
	config.IntVar(&params.redisPort, "redis-port", 6379, "Redis port")
	config.StringVar(&params.redisPassword, "redis-password", "", "Redis password")
}

func WithStore(fn func(store laboratory.Store) error) error {
	if params.useRedis {
		store, err := RedisStoreFactory(params.redisHost, params.redisPort, params.redisPassword)
		if err != nil {
			return err
		}

		return fn(store)
	} else {
		return fn(CsvStoreFactory(params.csvDir))
	}
}

var RedisStoreFactory = func(host string, port int, password string) (laboratory.Store, error) {
	store, err := NewRedisStore(host, port, password)
	if err != nil {
		return nil, err
	}

	return store, nil
}

var CsvStoreFactory = func(dir string) laboratory.Store {
	return NewCsvStore(dir)
}
