package services

import (
	"context"
	"log"
	"os"

	"go_project/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func EnsurePasswordHistoryIndex() {
	context := context.Background()
	db := config.MongoClient.Database(os.Getenv("MONGO_DATABASE"))
	collection := db.Collection("password_history")

	model := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "changed_at", Value: 1}},
		Options: options.Index().SetName("user_id_changed_at_index"),
	}

	_, err := collection.Indexes().CreateOne(context, model)
	if err != nil {
		log.Printf("⚠️ Failed to create password_history index: %v", err)
	} else {
		log.Println("✔️ Successfully created password_history index: user_id_changed_at_index")
	}

}
