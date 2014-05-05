package store

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cloudfoundry-incubator/pat/logs"
	"github.com/cloudfoundry-incubator/pat/redis"
)

type MetaStore struct {
	Directory string
	UseRedis  bool
	Conn      redis.Conn
}

type MetaData struct {
	StartTime   string `json:"start time"`
	Concurrency int
	Iterations  int
	Interval    int
	Stop        int
	Workload    string
	Note        string
}

func newMetaStore(directory string, useRedis bool) (*MetaStore, error) {
	if useRedis {
		var c redis.Conn
		err := redis.WithRedisConnection(func(conn redis.Conn) error {
			c = conn
			return nil
		})
		if err != nil {
			return nil, err
		}

		return &MetaStore{directory, useRedis, c}, nil
	} else {
		return &MetaStore{directory, useRedis, nil}, nil
	}
}

func (self *MetaStore) Write(fileName string, concurrency int, iterations int, interval int, stop int, workload string, note string) error {
	meta_data := MetaData{time.Now().Format(time.RFC850), concurrency, iterations, interval, stop, workload, note}

	if self.UseRedis {
		return self.writeRedis(fileName, meta_data)
	} else {
		return self.writeLocal(path.Join(self.Directory, fileName+".meta"), meta_data)
	}
}

func (self *MetaStore) writeLocal(outputPath string, meta_data MetaData) error {
	var logger = logs.NewLogger("store.meta")

	file, err := os.Create(outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Infof("Creating directory, %s", filepath.Dir(outputPath))
			os.MkdirAll(filepath.Dir(outputPath), 0755)
			file, err = os.Create(outputPath)
		}

		if err != nil {
			logger.Errorf("Can't write Meta: %v", err)
			return err
		}
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	header := []string{"start time", "concurrency", "iterations", "stop", "interval", "workload", "note"}
	body := []string{meta_data.StartTime, strconv.Itoa(meta_data.Concurrency),
		strconv.Itoa(meta_data.Iterations), strconv.Itoa(meta_data.Stop),
		strconv.Itoa(meta_data.Interval), meta_data.Workload, meta_data.Note}

	writer.Write(header)
	writer.Write(body)
	writer.Flush()

	return nil
}

func (self *MetaStore) writeRedis(name string, meta_data MetaData) error {
	var logger = logs.NewLogger("redis store.meta")

	json, err := json.Marshal(meta_data)
	if err != nil {
		logger.Errorf("Can't marshal json: %v", err)
		return err
	}

	_, err = self.Conn.Do("RPUSH", "experiment."+name+".meta", json)
	if err != nil {
		logger.Errorf("cannot push to redis: %v", err)
		return err
	}

	return nil
}
