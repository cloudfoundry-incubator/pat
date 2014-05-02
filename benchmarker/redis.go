package benchmarker

import (
	"encoding/json"
	"strings"
	"strconv"
	"net/url"
	
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

const DefaultTimeout = 60 * 5

func NewRedisWorker(conn redis.Conn) Worker {
	return NewRedisWorkerWithTimeout(conn, DefaultTimeout)
}

func NewRedisWorkerWithTimeout(conn redis.Conn, timeoutInSeconds int) Worker {
	return &rw{defaultWorker{make(map[string]workloads.WorkloadStep)}, conn, timeoutInSeconds}
}

func (rw rw) Time(workload string, workloadCtx map[string]interface{}) (result IterationResult) {
	guid, _ := uuid.NewV4()	
	workloadCtx = replaceSpaceWithEscape(workloadCtx)
	ctxContentStr, ctxTypeStr, ctxKeyStr := turnContextToRedisStr(workloadCtx)

	redisStr := "replies-"+guid.String() + " " + ctxTypeStr + " " + ctxKeyStr + " " + ctxContentStr + " " + workload
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
		_, err = slave.conn.Do("BLPOP", "stopped-"+slave.guid, DefaultTimeout)
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
			parts := strings.SplitN(reply[1], " ", 5)

			workloadCtx := fillContextFromRedisStr(parts[1], parts[2], parts[3])

			go func(experiment string, replyTo string, workloadCtx map[string]interface{}) {				
				result := delegate.Time(experiment, workloadCtx)				
				var encoded []byte
				encoded, err = json.Marshal(result)
				logger.Debug("Completed slave task, replying")
				conn.Do("RPUSH", replyTo, string(encoded))
			}(parts[4], parts[0], workloadCtx)
		}

		if err != nil {
			logger.Warnf("ERROR: slave encountered error: %v", err)
		}
	}
}

func turnContextToRedisStr(workloadCtx map[string]interface{})(strValue string, strType string, strKey string){	
	for k, v := range workloadCtx {
		if (strValue != "") { strValue = strValue + "," }
		if (strType != "") { strType = strType + "," }
		if (strKey != "") { strKey = strKey + "," }
	
		strKey = strKey + k

		switch v.(type) {
			default:
				logger := logs.NewLogger("redis")
				logger.Error("Unsupported type in context map")
			case int:
				strType = strType + "int"				
				strValue = strValue + strconv.Itoa(v.(int))
			case int64:
				strType = strType + "int64"	
				strValue = strValue + strconv.FormatInt(v.(int64), 10)
			case string:
				strType = strType + "string"
				strValue = strValue + v.(string)
			case bool:
				strType = strType + "bool"
				strValue = strValue + strconv.FormatBool(v.(bool))
		}
	}
	return strValue, strType, strKey
}

func fillContextFromRedisStr(ctxType string, ctxKey string, ctxContent string) map[string]interface{} {
	var intStr int64
	
	workloadCtx := make(map[string]interface{})
	contentList := strings.Split(ctxContent, ",")
	keyList := strings.Split(ctxKey, ",")

	for i, v := range strings.Split(ctxType,",") {
		switch {
		case v == "string":
			workloadCtx[keyList[i]], _ = url.QueryUnescape(contentList[i])
			break
		case v == "int":		
			intStr, _ = strconv.ParseInt(contentList[i], 0, 0)
			workloadCtx[keyList[i]] = int(intStr)
			break
		case v == "int64":
			workloadCtx[keyList[i]], _ = strconv.ParseInt(contentList[i], 0, 64)
			break
		case v == "bool":			
			workloadCtx[keyList[i]], _ =  strconv.ParseBool(contentList[i])
			break
		}
	}
	return workloadCtx
}

func replaceSpaceWithEscape(workloadCtx map[string]interface{}) map[string]interface{} {
	for k, _ := range workloadCtx {
		if (reflect.TypeOf(workloadCtx[k]).Name() == "string") {
			workloadCtx[k] = url.QueryEscape(workloadCtx[k].(string))
		}
	}
	return workloadCtx
}
