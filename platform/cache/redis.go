package cache

import (
	"os"
	"strconv"

	"github.com/Komentory/utilities"
	"github.com/go-redis/redis/v8"
)

// RedisConnection func for connect to Redis server.
func RedisConnection() (*redis.Client, error) {
	// Define Redis database number.
	dbNumber, err := strconv.Atoi(os.Getenv("REDIS_DB_NUMBER"))
	if err != nil {
		return nil, err
	}

	// Build Redis connection URL.
	redisConnURL, err := utilities.ConnectionURLBuilder("redis")
	if err != nil {
		return nil, err
	}

	// Return Redis client with options.
	return redis.NewClient(
		&redis.Options{
			Addr:     redisConnURL,
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       dbNumber,
		},
	), nil
}
