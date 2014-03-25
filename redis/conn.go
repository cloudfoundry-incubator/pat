package redis

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
)

const MAX_IDLE = 20

type Conn interface {
	Do(cmd string, args ...interface{}) (interface{}, error)
}

type conn struct {
	p *redis.Pool
}

func Connect(host string, port int, password string) (Conn, error) {
	fmt.Printf("Dialling redis on %s:%d", host, port)
	return test(&conn{redis.NewPool(func() (redis.Conn, error) { return connect(host, port, password) }, MAX_IDLE)}, nil)
}

func connect(host string, port int, password string) (redis.Conn, error) {
	r, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	if password != "" {
		auth(r, password)
	}

	return r, nil
}

func test(conn *conn, err error) (Conn, error) {
	if err == nil {
		_, err = conn.Do("LRANGE", "test", 0, 0)
	}

	if err != nil {
		fmt.Println("Redis connection failed")
	}
	return conn, err
}

func auth(c redis.Conn, password string) error {
	if _, err := c.Do("AUTH", password); err != nil {
		c.Close()
		return err
	}

	return nil
}

func (c conn) Do(cmd string, args ...interface{}) (interface{}, error) {
	return c.withConnection(func(c redis.Conn) (interface{}, error) {
		return c.Do(cmd, args...)
	})
}

func (c conn) withConnection(fn func(c redis.Conn) (interface{}, error)) (interface{}, error) {
	conn := c.p.Get()
	defer conn.Close()
	return fn(conn)
}

func String(reply interface{}, err error) (string, error) {
	return redis.String(reply, err)
}

func Strings(reply interface{}, err error) ([]string, error) {
	return redis.Strings(reply, err)
}

func Bytes(reply interface{}, err error) ([]byte, error) {
	return redis.Bytes(reply, err)
}
