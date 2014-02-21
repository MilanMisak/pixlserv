package main

import (
	"io"
	"log"

	"github.com/garyburd/redigo/redis"
)

func cacheInit() error {
	// TODO - make the port configurable?
	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		return err
	}
	defer c.Close()

	log.Println("Cache ready")
	return nil
}

// Adds the given file to the cache.
func addToCache(filePath string, w io.Writer) {
	// TODO - implement
}

// Checks if the given path is in the cache.
func fileExistsInCache(filePath string) bool {
	// TODO - implement
	return false
}
