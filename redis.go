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
	conn redis.Conn
)

func redisInit() error {
	port, err := strconv.Atoi(os.Getenv(redisPortEnvVar))
	if err != nil {
		port = redisDefaultPort
	}
	conn, err = redis.Dial("tcp", ":"+strconv.Itoa(port))
	return err
}
