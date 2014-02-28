package main

import (
	"errors"
	"image"
	"log"
	"os"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

const (
	REDIS_PORT_ENV_VAR = "PIXLSERV_REDIS_PORT"
	REDIS_DEFAULT_PORT = 6379
)

func cacheInit() error {
	port, err := strconv.Atoi(os.Getenv(REDIS_PORT_ENV_VAR))
	if err != nil {
		port = REDIS_DEFAULT_PORT
	}

	c, err := redis.Dial("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	defer c.Close()

	log.Printf("Cache ready, using port %d", port)
	return nil
}

// Adds the given file to the cache.
func addToCache(filePath string, img image.Image, format string) error {
	// TODO - implement
	log.Println("Adding to cache:", filePath)

	return saveImage(img, format, filePath)
}

// Checks if the given path is in the cache.
func loadFromCache(filePath string) (image.Image, string, error) {
	// TODO - implement
	log.Println("Checking for file:", filePath)
	//return false

	// TODO - update last accessed flag

	//img, format, err := loadImage(fullImagePath)
	return nil, "", errors.New("Image not found")
}
