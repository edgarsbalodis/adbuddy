package storage

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	Client *mongo.Client
}

func NewMongoClient(connection string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connection))
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Connected to Database")
	}
	return client, nil
}

func New(client *mongo.Client) *Storage {
	return &Storage{
		Client: client,
	}
}

func (s *Storage) GetCollection(collection string) *mongo.Collection {
	coll := s.Client.Database("adbuddy").Collection(collection)
	return coll
}
