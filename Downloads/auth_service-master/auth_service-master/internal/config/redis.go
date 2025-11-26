package config

import (
    "os"
    "time"

    "github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
    addr := os.Getenv("REDIS_ADDR")
    if addr == "" {
        addr = "localhost:6379"
    }

    return redis.NewClient(&redis.Options{
        Addr:         addr,
        Password:     "",
        DB:           0,
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
    })
}
