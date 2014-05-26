package main

import (
	"os"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

const (
	redisPortEnvVar  = "PIXLSERV_REDIS_PORT"
	redisUrlEnvVar   = "PIXLSERV_REDIS_URL"
	redisDefaultPort = 6379
)

var (
	Conn redis.Conn
)

func redisInit() error {
	url := os.Getenv(redisUrlEnvVar)
	var err error
	if url != "" {
		Conn, err = redis.Dial("tcp", url)
		return err
	} else {
		port, err := strconv.Atoi(os.Getenv(redisPortEnvVar))
		if err != nil {
			port = redisDefaultPort
		}

		Conn, err = redis.Dial("tcp", ":"+strconv.Itoa(port))
		return err
	}
}

func redisCleanUp() {
	Conn.Close()
}
