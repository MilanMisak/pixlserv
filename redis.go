package main

import (
	"os"
	"strconv"

	"github.com/garyburd/redigo/redis"
	"github.com/soveran/redisurl"
)

const (
	redisPortEnvVar  = "PIXLSERV_REDIS_PORT"
	redisURLEnvVar   = "PIXLSERV_REDIS_URL"
	redisDefaultPort = 6379
)

var (
	// Conn is a global redis connection object
	Conn redis.Conn
)

func redisInit() error {
	url := os.Getenv(redisURLEnvVar)
	var err error
	if url != "" {
		Conn, err = redisurl.ConnectToURL(url)
	} else {
		port, err := strconv.Atoi(os.Getenv(redisPortEnvVar))
		if err != nil {
			port = redisDefaultPort
		}

		Conn, err = redis.Dial("tcp", ":"+strconv.Itoa(port))
	}
	return err
}

func redisCleanUp() {
	Conn.Close()
}
