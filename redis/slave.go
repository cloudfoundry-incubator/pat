package redis

import (
	"github.com/garyburd/redigo/redis"
	"strings"
	"time"
)

type Slave struct {
	in          Connection
	out         Connection
	channel     string
	experiments map[string]func() (time.Duration, error)
}

func NewSlave(in Connection, out Connection, channel string) *Slave {
	return &Slave{in, out, channel, make(map[string]func() (time.Duration, error))}
}

func (self *Slave) WithExperiment(name string, fn func() (time.Duration, error)) *Slave {
	self.experiments[name] = fn
	return self
}

func (self *Slave) Next() error {
	message, err := redis.String(self.in.Do("BLPOP", self.channel))
	if err != nil {
		return err
	}

	task := strings.SplitN(message, ",", 2)

	duration, _ := self.experiments[task[1]]()
	self.out.Do("RPUSH", task[0], duration)
	return nil
}
