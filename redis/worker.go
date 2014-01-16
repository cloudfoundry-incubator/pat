package redis

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

type Connection interface {
	Do(commandName string, args ...interface{}) (reply interface{}, err error)
}

type Worker struct {
	out           Connection // NOTE(jz) - probably have to make sure this is pooled and not shared
	in            Connection // NOTE(jz) - probably have to make sure this is pooled and not shared
	channel       string
	reply_channel string
}

func NewWorker(out Connection, in Connection, channel string, reply_channel string) *Worker {
	return &Worker{out, in, channel, reply_channel}
}

func (self *Worker) Time(experiment string) (time.Duration, error) {
	self.out.Do("RPUSH", self.channel, self.reply_channel, experiment)
	nanos, err := redis.Int64(self.in.Do("BLPOP", self.reply_channel))
	return time.Duration(nanos), err
}
