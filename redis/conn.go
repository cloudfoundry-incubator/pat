package redis

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
)

type Conn interface {
	Do(cmd string, args ...interface{}) (interface{}, error)
}

type conn struct {
	c redis.Conn
}

func Connect(host string, port int, password string) (Conn, error) {
	r, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	if password != "" {
		auth(r, password)
	}

	return &conn{r}, nil
}

func auth(c redis.Conn, password string) error {
	if _, err := c.Do("AUTH", password); err != nil {
		c.Close()
		return err
	}

	return nil
}

func (c conn) Do(cmd string, args ...interface{}) (interface{}, error) {
	return c.c.Do(cmd, args...)
}

func Strings(reply interface{}, err error) ([]string, error) {
	return redis.Strings(reply, err)
}
