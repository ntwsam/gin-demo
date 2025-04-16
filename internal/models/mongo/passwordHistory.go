package mongoModel

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PasswordHistory struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    string             `bson:"user_id"`
	Password  string             `bson:"password"`
	ChangedAt time.Time          `bson:"changed_at"`
}
