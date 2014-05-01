package benchmarker

import (
	"encoding/json"
	"strings"
	"strconv"

	"github.com/cloudfoundry-incubator/pat/logs"
	"github.com/cloudfoundry-incubator/pat/redis"
	"github.com/cloudfoundry-incubator/pat/workloads"
	"github.com/nu7hatch/gouuid"
)

type rw struct {
	defaultWorker
	conn             redis.Conn
	timeoutInSeconds int
}

const DEFAULT_TIMEOUT = 60 * 5

func NewRedisWorker(conn redis.Conn) Worker {
	return NewRedisWorkerWithTimeout(conn, DEFAULT_TIMEOUT)
}

func NewRedisWorkerWithTimeout(conn redis.Conn, timeoutInSeconds int) Worker {
	return &rw{defaultWorker{make(map[string]workloads.WorkloadStep)}, conn, timeoutInSeconds}
}

func (rw rw) Time(experiment string, workerIndex int) (result IterationResult) {
	guid, _ := uuid.NewV4()
	rw.conn.Do("RPUSH", "tasks", "replies-"+guid.String()+" "+strconv.Itoa(workerIndex)+" "+experiment)	
	reply, err := redis.Strings(rw.conn.Do("BLPOP", "replies-"+guid.String(), rw.timeoutInSeconds))

	if err != nil {
		return IterationResult{0, []StepResult{}, encodeError(err)}
	} else {
		json.Unmarshal([]byte(reply[1]), &result)
		return
	}
}

type slave struct {
	guid string
	conn redis.Conn
}

func StartSlave(conn redis.Conn, delegate Worker) slave {
	guid, _ := uuid.NewV4()
	go slaveLoop(conn, delegate, guid.String())
	return slave{guid.String(), conn}
}

func (slave slave) Close() error {
	_, err := slave.conn.Do("RPUSH", "stop-"+slave.guid, true)
	if err == nil {
		_, err = slave.conn.Do("BLPOP", "stopped-"+slave.guid, DEFAULT_TIMEOUT)
	}

	logs.NewLogger("redis.slave").Infof("Redis slave shutting down, %v", err)
	return err
}

func slaveLoop(conn redis.Conn, delegate Worker, handle string) {
	logger := logs.NewLogger("redis.slave")
	logger.Info("Started slave")

	for {
		reply, err := redis.Strings(conn.Do("BLPOP", "stop-"+handle, "tasks", 0))

		if len(reply) == 0 {
			panic("Empty task, usually means connection lost, shutting down")
		}

		if reply[0] == "stop-"+handle {
			conn.Do("RPUSH", "stopped-"+handle, true)
			break
		}

		if err == nil {
			parts := strings.SplitN(reply[1], " ", 3)

			go func(experiment string, replyTo string, workerIndex string) {
				i, _ := strconv.Atoi(workerIndex)
				result := delegate.Time(experiment, i)
				var encoded []byte
				encoded, err = json.Marshal(result)
				logger.Debug("Completed slave task, replying")
				conn.Do("RPUSH", replyTo, string(encoded))
			}(parts[2], parts[0], parts[1])
		}

		if err != nil {
			logger.Warnf("ERROR: slave encountered error: %v", err)
		}
	}
}
