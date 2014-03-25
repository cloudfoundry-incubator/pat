package redis

import (
	"encoding/json"

	"github.com/cloudfoundry-community/pat/config"
)

var params = struct {
	redisHost     string
	redisPort     int
	redisPassword string
	vcapServices  string
}{}

func DescribeParameters(config config.Config) {
	config.StringVar(&params.redisHost, "redis-host", "localhost", "Redis hostname")
	config.IntVar(&params.redisPort, "redis-port", 6379, "Redis port")
	config.StringVar(&params.redisPassword, "redis-password", "", "Redis password")
	config.EnvVar(&params.vcapServices, "VCAP_SERVICES", "", "The VCAP_SERVICES environment variable")
}

func WithRedis(fn func(conn Conn) error) error {
	parseVcapServices()
	store, err := ConnFactory(params.redisHost, params.redisPort, params.redisPassword)
	if err == nil {
		err = fn(store)
	}

	return err
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

var ConnFactory = func(host string, port int, password string) (Conn, error) {
	return Connect(host, port, password)
}
