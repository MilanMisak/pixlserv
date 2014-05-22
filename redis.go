package main

import (
	"os"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

const (
	redisPortEnvVar  = "PIXLSERV_REDIS_PORT"
	redisDefaultPort = 6379
)

var (
	Conn redis.Conn
)

func redisInit() error {
	port, err := strconv.Atoi(os.Getenv(redisPortEnvVar))
	if err != nil {
		port = redisDefaultPort
	}
	Conn, err = redis.Dial("tcp", ":"+strconv.Itoa(port))
	return err
}

func redisCleanUp() {
	Conn.Close()
}
