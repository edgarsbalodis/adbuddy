package storage

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Filter struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty"`
	Name      string               `bson:"name"`
	Value     string               `bson:"value"`
	Questions []primitive.ObjectID `bson:"questions"`
}

func (s *Storage) GetFilters() []Filter {
	coll := s.GetCollection("filters")

	cursor, err := coll.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	// Decode each filter from the cursor
	var filters []Filter
	if err = cursor.All(context.TODO(), &filters); err != nil {
		log.Fatal(err)
	}

	// Now, 'records' slice contains all the records from the collection.
	for _, filter := range filters {
		fmt.Println(filter)
	}

	return filters
}
