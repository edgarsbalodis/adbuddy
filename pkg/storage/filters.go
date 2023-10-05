package storage

import "go.mongodb.org/mongo-driver/bson/primitive"

type Filters struct {
	ID        primitive.ObjectID   `bson:"_id, omitempty"`
	Name      string               `bson:"name"`
	Questions []primitive.ObjectID `bson:"questions"`
}
