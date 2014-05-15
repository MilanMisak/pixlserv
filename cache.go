package main

import (
	"errors"
	"fmt"
	"image"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	redisPortEnvVar  = "PIXLSERV_REDIS_PORT"
	redisDefaultPort = 6379
)

var (
	conn redis.Conn
)

func cacheInit() error {
	port, err := strconv.Atoi(os.Getenv(redisPortEnvVar))
	if err != nil {
		port = redisDefaultPort
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
	size, err := saveImage(img, format, filePath)
	if err == nil {
		key := fmt.Sprintf("image:%s", filePath)

		// Add a record to the cache
		conn.Do("HSET", key, "size", size)

		conn.Do("SETNX", "totalcachesize", 0)
		conn.Do("INCRBY", "totalcachesize", size)

		// Update queue of last accesses
		cacheUpdateLastAccess(key)

		pruneCache()
	}

	return err
}

func removeFromCache(key string) {
	log.Printf("Removing from cache: %s", key)

	size, err := redis.Int(conn.Do("HGET", key, "size"))
	if err != nil {
		return
	}

	err = deleteImage(strings.Replace(key, "image:", "", 1))
	if err != nil {
		log.Println("Error removing image:", err)
		return
	}
	conn.Do("DEL", key)
	conn.Do("ZREM", "imageaccesses", key)
	conn.Do("DECRBY", "totalcachesize", size)
}

// Loads a file specified by its path from the cache.
func loadFromCache(filePath string) (image.Image, string, error) {
	log.Println("Cache lookup for:", filePath)

	exists, err := redis.Bool(conn.Do("EXISTS", fmt.Sprintf("image:%s", filePath)))
	if err != nil {
		return nil, "", err
	}

	if exists {
		key := fmt.Sprintf("image:%s", filePath)
		cacheUpdateLastAccess(key)

		return loadImage(filePath)
	}

	return nil, "", errors.New("image not found")
}

func cacheUpdateLastAccess(key string) {
	timestamp := time.Now().Unix()
	conn.Do("ZADD", "imageaccesses", timestamp, key)
}

func pruneCache() {
	go func() {
		config := getConfig()
		if config.cacheLimit == 0 {
			return
		}

		totalCacheSize, err := redis.Int(conn.Do("GET", "totalcachesize"))
		if err != nil {
			return
		}

		log.Printf("total size: %d", totalCacheSize)

		if totalCacheSize < config.cacheLimit {
			return
		}

		candidate := getCacheRemovalCandidate()
		if candidate != "" {
			removeFromCache(candidate)
		}
	}()
}

// TODO - return multiple to speed things up
func getCacheRemovalCandidate() string {
	//config := getConfig()

	// LRU
	candidates, err := redis.Strings(conn.Do("ZRANGE", "imageaccesses", 0, 0))
	if err == nil && len(candidates) > 0 {
		return candidates[0]
	}
	return ""
}
