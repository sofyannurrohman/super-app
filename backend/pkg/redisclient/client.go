package redisclient

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

// Connect initializes the global Redis client.
func Connect() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	password := os.Getenv("REDIS_PASSWORD")

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,

		// Connection pool tuning — cegah bottleneck di concurrent requests
		PoolSize:        30,              // naikkan dari default 10
		MinIdleConns:    5,               // selalu siapkan 5 koneksi siap pakai
		ConnMaxIdleTime: 5 * time.Minute, // bebaskan koneksi idle setelah 5 menit
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
	})

	ctx := context.Background()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Printf("[Redis] WARNING: could not connect to Redis at %s: %v", addr, err)
	} else {
		log.Printf("[Redis] Connected to Redis at %s", addr)
	}

	Client = rdb
	return rdb
}
