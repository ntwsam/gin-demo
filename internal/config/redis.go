package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectRedis() {

	// 🪸 ดึงข้อมูลจาก .env
	redisAddress := os.Getenv("REDIS_ADDRESS")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// 🪸 สร้าง client ไว้เชื่อมต่อ
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: redisPassword,
		DB:       0, // 🐳 default
	})

	// 🪸 เชื่อมต่อและเช็คสถานะ
	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("⚠️ Failed connecting to Redis: %v", err)
	}
	fmt.Println("✔️ Redis connection successful!")
}
