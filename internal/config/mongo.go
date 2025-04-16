package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

func ConnectMongoDB() {

	// 🪸 ดึงข้อมูลจาก .env
	mongoURI := os.Getenv("MONGO_URI")

	// 🪸 สร้าง client และกำหนด uri
	clientOptions := options.Client().ApplyURI(mongoURI)

	// 🪸 เชื่อมต่อ mongo
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("⚠️ Failed connecting to MongoDB: %v", err)
	}

	// 🪸 ตรวจสอบการเชื่อมต่อ
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("⚠️ Failed to ping MongoDB: %v", err)
	}
	fmt.Println("✔️ Redis connection successful!")

	MongoClient = client
}
