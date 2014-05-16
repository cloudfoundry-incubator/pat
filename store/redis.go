package store

import (
	"time"
	"strconv"
	"encoding/json"

	"github.com/cloudfoundry-incubator/pat/experiment"
	"github.com/cloudfoundry-incubator/pat/redis"
)

const MAX_RESULTS = 10000

type MetaData struct {
	Guid                string
	StartTime           string `json:"start time"`
	Concurrency         string
	ConcurrencyStepTime string `json:"concurrency step time"`
	Iterations          int
	Interval            int
	Stop                int
	Workload            string
	Note                string
}

type redisStore struct {
	c redis.Conn
}

type redisExperiment struct {
	redisStore *redisStore
	guid       string
}

func NewRedisStore(conn redis.Conn) (*redisStore, error) {
	return &redisStore{conn}, nil
}

func (r *redisStore) LoadAll() ([]experiment.Experiment, error) {
	c := r.c
	members, err := redis.Strings(c.Do("LRANGE", "experiments", 0, MAX_RESULTS))
	if err != nil {
		return nil, err
	}

	experiments := make([]experiment.Experiment, len(members))
	for i := 0; i < len(members); i++ {
		experiments[i] = &redisExperiment{r, members[i]}
	}

	return experiments, nil
}

func (r *redisStore) Writer(guid string, ex experiment.ExperimentConfiguration) func(samples <-chan *experiment.Sample) {
	r.c.Do("RPUSH", "experiments", guid)
	r.writeMetaData(guid, ex)
	return func(ch <-chan *experiment.Sample) {
		for sample := range ch {
			push(r.c, guid, sample)
		}
	}
}

func (r *redisStore) writeMetaData(guid string, ex experiment.ExperimentConfiguration) {
	var concurrency string
	for iter, value := range ex.Concurrency {
		if iter >= 1 {
			concurrency += ".."  + strconv.Itoa(value)
		} else {
			concurrency += strconv.Itoa(value)
		}
	}

	metaData := MetaData{Guid: guid, StartTime: time.Now().Format(time.RFC850), Concurrency: concurrency,
			ConcurrencyStepTime: ex.ConcurrencyStepTime.String(), Iterations: ex.Iterations,
			Interval: ex.Interval, Stop: ex.Stop, Workload: ex.Workload, Note: ex.Note}

	json, err := json.Marshal(metaData)
	if err != nil {
		return
	}

	_, err = r.c.Do("RPUSH", "meta_data", json)
	if err != nil {
		return
	}
}

func push(c redis.Conn, guid string, sample *experiment.Sample) {
	json, _ := json.Marshal(sample)
	c.Do("RPUSH", "experiment."+guid, json)
}

func (r redisExperiment) GetData() ([]*experiment.Sample, error) {
	members, err := redis.Strings(r.redisStore.c.Do("LRANGE", "experiment."+r.guid, 0, MAX_RESULTS))
	if err != nil {
		return nil, err
	}

	samples := make([]*experiment.Sample, len(members))
	for i := 0; i < len(samples); i++ {
		m := members[i]
		b := []byte(m)
		json.Unmarshal(b, &samples[i])
	}

	return samples, nil
}

func (r redisExperiment) GetGuid() string {
	return r.guid
}
