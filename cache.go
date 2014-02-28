package main

import (
	"errors"
	"fmt"
	"image"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	REDIS_PORT_ENV_VAR = "PIXLSERV_REDIS_PORT"
	REDIS_DEFAULT_PORT = 6379
)

var (
	conn redis.Conn = nil
)

func cacheInit() error {
	port, err := strconv.Atoi(os.Getenv(REDIS_PORT_ENV_VAR))
	if err != nil {
		port = REDIS_DEFAULT_PORT
	}

	conn, err = redis.Dial("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}

	log.Printf("Cache ready, using port %d", port)

	return nil
}

func cacheCleanUp() {
	log.Println("Closing redis connection for the cache")
	conn.Close()
}

// Adds the given file to the cache.
func addToCache(filePath string, img image.Image, format string) error {
	log.Println("Adding to cache:", filePath)

	// Save the image
	err := saveImage(img, format, filePath)
	if err == nil {
		// Add a record to the cache
		timestamp, _ := time.Now().MarshalText()
		conn.Do("SET", fmt.Sprintf("image:%s", filePath), timestamp)
	}

	return err
}

// Loads a file specified by its path from the cache.
func loadFromCache(filePath string) (image.Image, string, error) {
	log.Println("Cache lookup for:", filePath)

	exists, err := redis.Bool(conn.Do("EXISTS", fmt.Sprintf("image:%s", filePath)))
	if err != nil {
		return nil, "", err
	}

	if exists {
		// TODO - update last accessed flag
		return loadImage(filePath)
	}

	return nil, "", errors.New("Image not found")
}
