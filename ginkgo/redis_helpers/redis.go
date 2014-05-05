package redis_helpers

import (
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"time"
)

func StartRedis(config string) {
	_, filename, _, _ := runtime.Caller(0)
	dir, _ := filepath.Abs(filepath.Dir(filename))
	StopRedis()
	exec.Command("redis-server", path.Join(dir, config)).Run()
	time.Sleep(450 * time.Millisecond) // yuck(jz)
}

func StopRedis() {
	exec.Command("redis-cli", "-p", "63798", "shutdown").Run()
	exec.Command("redis-cli", "-p", "63798", "-a", "p4ssw0rd", "shutdown").Run()
}

func StopLocalRedis() {
	exec.Command("redis-cli", "shutdown").Run()
}

func CheckRedisRunning() (err error) {
	for i := 0; i < 20; i++ {
		err = exec.Command("redis-cli", "ping").Run()
		if err == nil {
			return err
		}

		time.Sleep(500 * time.Millisecond)
	}

	return err
}
