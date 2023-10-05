package storage

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	client *mongo.Client
}

type Response struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    int64              `bson:"userID"`
	ChatID    int64              `bson:"chatID"`
	Type      string             `bson:"type"`
	Timestamp time.Time          `bson:"timestamp"`
	Answers   []Answer           `bson:"answers"`
}

type Answer struct {
	Key   string `bson:"key"`
	Value string `bson:"value"`
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

func (db *DB) GetCollection(collection string) *mongo.Collection {
	coll := db.client.Database("adbuddy").Collection(collection)
	return coll
}
