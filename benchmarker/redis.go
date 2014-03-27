package benchmarker

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudfoundry-community/pat/redis"
	"github.com/cloudfoundry-community/pat/workloads"
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

func (rw rw) Time(experiment string) (result IterationResult) {
	guid, _ := uuid.NewV4()
	rw.conn.Do("RPUSH", "tasks", "replies-"+guid.String()+" "+experiment)
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

func StartSlave(conn redis.Conn, worker *LocalWorker) slave {
	guid, _ := uuid.NewV4()
	go slaveLoop(conn, worker, guid.String())
	return slave{guid.String(), conn}
}

func (slave slave) Close() error {
	_, err := slave.conn.Do("RPUSH", "stop-"+slave.guid, true)
	if err == nil {
		_, err = slave.conn.Do("BLPOP", "stopped-"+slave.guid, DEFAULT_TIMEOUT)
	}

	fmt.Println("Redis slave shutting down", err)
	return err
}

func slaveLoop(conn redis.Conn, worker *LocalWorker, handle string) {
	fmt.Println("Started slave")
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
			parts := strings.SplitN(reply[1], " ", 2)
			go func(experiment string, replyTo string) {
				result := worker.Time(experiment)
				var encoded []byte
				encoded, err = json.Marshal(result)
				fmt.Println("Completed slave task, replying")
				conn.Do("RPUSH", replyTo, string(encoded))
			}(parts[1], parts[0])
		}

		if err != nil {
			fmt.Println("ERROR: slave encountered error: ", err)
		}
	}
}
