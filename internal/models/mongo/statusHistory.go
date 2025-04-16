package mongoModel

import "go.mongodb.org/mongo-driver/bson/primitive"

type StatusHistory struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	UserID string             `bson:"user_id"`
	Status string             `bson:"status"`
}
