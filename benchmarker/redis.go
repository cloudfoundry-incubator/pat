package benchmarker

import (
	"encoding/json"
	"strings"
	"strconv"

	"github.com/cloudfoundry-community/pat/logs"
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

const spaceEscapeStr = "<%>"

func NewRedisWorker(conn redis.Conn) Worker {
	return NewRedisWorkerWithTimeout(conn, DEFAULT_TIMEOUT)
}

func NewRedisWorkerWithTimeout(conn redis.Conn, timeoutInSeconds int) Worker {
	return &rw{defaultWorker{make(map[string]workloads.WorkloadStep)}, conn, timeoutInSeconds}
}

func (rw rw) Time(workload string, workloadCtx map[string]interface{}) (result IterationResult) {
	guid, _ := uuid.NewV4()	
	workloadCtx = initNilContextMap(workloadCtx)
	escapeStr := generateEscapeStr(workloadCtx)
	workloadCtx = replaceSpaceWithEscape(workloadCtx, escapeStr)
	ctxStr := concatContextStr(workloadCtx)

	if (workloadCtx["workerIndex"] == nil) { workloadCtx["workerIndex"] = 0 }
	workerIndex := strconv.Itoa(workloadCtx["workerIndex"].(int))
	
	redisStr := escapeStr + " replies-"+guid.String() + " " + workerIndex + ctxStr + " " + workload
	rw.conn.Do("RPUSH", "tasks", redisStr)

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
			parts := strings.SplitN(reply[1], " ", 4 + len(RedisContextMapStr))
			escapeStr := parts[0]
			workloadCtx := make(map[string]interface{})
			workerIndex, _ := strconv.Atoi(parts[2])
			workloadCtx["workerIndex"] = workerIndex
			for i, v := range RedisContextMapStr {
				workloadCtx[v] = parts[(i+3)]				
				workloadCtx[v] = strings.Replace(workloadCtx[v].(string), escapeStr, " ", -1)
			}

			go func(experiment string, replyTo string, workloadCtx map[string]interface{}) {				
				result := delegate.Time(experiment, workloadCtx)				
				var encoded []byte
				encoded, err = json.Marshal(result)
				logger.Debug("Completed slave task, replying")
				conn.Do("RPUSH", replyTo, string(encoded))
			}(parts[6], parts[1], workloadCtx)
		}

		if err != nil {
			logger.Warnf("ERROR: slave encountered error: %v", err)
		}
	}
}

func replaceSpaceWithEscape(workloadCtx map[string]interface{}, escapeStr string) map[string]interface{} {
	for _, v := range RedisContextMapStr {
		workloadCtx[v] = strings.Replace(workloadCtx[v].(string), " ", escapeStr, -1)				
	}
	return workloadCtx
}

func generateEscapeStr(workloadCtx map[string]interface{}) string {
	escapeStr := ""
	noEscape := false

	for noEscape != true {

		noEscape = true
		escapeStr = escapeStr + spaceEscapeStr

		//check to see context strings contain escapeStr
		for _, v := range RedisContextMapStr {
			if strings.Contains(workloadCtx[v].(string), escapeStr) {
				noEscape = false
			}
		}
			
	}

	return escapeStr
}

func concatContextStr(workloadCtx map[string]interface{}) string {
	ctxStr := ""

	for _, v := range RedisContextMapStr {
		if (workloadCtx[v] == nil) {
			workloadCtx[v] = ""		
		}
		ctxStr = ctxStr + " " + workloadCtx[v].(string)
	}

	return ctxStr
}

func initNilContextMap(workloadCtx map[string]interface{}) map[string]interface{} { 
	for _, v := range RedisContextMapStr {
		if (workloadCtx[v] == nil) {
			workloadCtx[v] = ""
		}
	}
	return workloadCtx
}
