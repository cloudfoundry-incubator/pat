package store

import (
	"encoding/json"

	"github.com/julz/pat/config"
	"github.com/julz/pat/laboratory"
)

var params = struct {
	csvDir        string
	useRedis      bool
	redisHost     string
	redisPort     int
	redisPassword string
	vcapServices  string
}{}

func DescribeParameters(config config.Config) {
	config.StringVar(&params.csvDir, "csv-dir", "output/csvs", "Directory to Store CSVs")
	config.BoolVar(&params.useRedis, "use-redis", false, "True if redis should be used (requires the -redis-host, -redis-port and -redis-password arguments)")
	config.StringVar(&params.redisHost, "redis-host", "localhost", "Redis hostname")
	config.IntVar(&params.redisPort, "redis-port", 6379, "Redis port")
	config.StringVar(&params.redisPassword, "redis-password", "", "Redis password")
	config.EnvVar(&params.vcapServices, "VCAP_SERVICES", "", "The VCAP_SERVICES environment variable")
}

func WithStore(fn func(store laboratory.Store) error) error {
	parseVcapServices()
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

func parseVcapServices() {
	if params.vcapServices != "" {
		b := []byte(params.vcapServices)
		v := make(map[string][]struct {
			Name        string `json:"name"`
			Credentials struct {
				Hostname string `json:"hostname"`
				Port     int    `json:"port"`
				Password string `json:"password"`
			} `json:"credentials"`
		})

		json.Unmarshal(b, &v)
		for _, val := range v {
			firstMatch := val[0]
			if firstMatch.Name == "redis" {
				params.redisHost = firstMatch.Credentials.Hostname
				params.redisPort = firstMatch.Credentials.Port
				params.redisPassword = firstMatch.Credentials.Password
			}
		}
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
