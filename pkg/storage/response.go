package storage

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Response struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    int64              `bson:"userID"`
	ChatID    int64              `bson:"chatID"`
	Type      string             `bson:"type"`
	Timestamp time.Time          `bson:"timestamp"`
	Answers   []Answer           `bson:"answers"`
}

type Answer struct {
	Key   string      `bson:"key"`
	Value interface{} `bson:"value"`
}

func (s *Storage) SaveResponse(response Response) {
	coll := s.GetCollection("responses")
	// insert in DB
	_, err := coll.InsertOne(context.TODO(), response)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Storage) DeleteResponse(responseID string) {
	coll := s.GetCollection("responses")

	//delete from DB
	id, err := primitive.ObjectIDFromHex(responseID)
	if err != nil {
		log.Fatal(err)
	}

	filter := bson.M{"_id": id}

	res, err := coll.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(res)
}

func (s *Storage) UpdateTimestamp(responseID string) error {
	id, err := primitive.ObjectIDFromHex(responseID)
	if err != nil {
		return err
	}
	coll := s.GetCollection("responses")
	// update timestamp for filter
	f := bson.D{primitive.E{Key: "_id", Value: id}}
	loc, _ := time.LoadLocation("Europe/Riga")
	update := bson.D{{"$set", bson.D{{"timestamp", time.Now().In(loc)}}}}

	res, updateErr := coll.UpdateOne(context.TODO(), f, update)
	if updateErr != nil {
		return updateErr
	}

	log.Printf("Response updated: %v", res)
	return nil
}

func (s *Storage) GetResponse(responseID string) Response {
	id, err := primitive.ObjectIDFromHex(responseID)
	if err != nil {
		log.Println("Invalid id")
		return Response{}
	}
	coll := s.GetCollection("responses")
	result := coll.FindOne(context.Background(), bson.M{"_id": id})

	response := Response{}
	result.Decode(&response)
	loc, _ := time.LoadLocation("Europe/Riga")
	response.Timestamp = response.Timestamp.In(loc)
	return response
}

// returns all responses
func (s *Storage) GetResponses() []Response {
	coll := s.GetCollection("responses")

	result, err := coll.Find(context.TODO(), bson.D{{}})
	if err != nil {
		log.Printf("Error while finding responses: %v", err)
	}
	var data []Response
	if err := result.All(context.Background(), &data); err != nil {
		log.Fatal(err)
	}
	loc, _ := time.LoadLocation("Europe/Riga")
	for i := range data {
		data[i].Timestamp = data[i].Timestamp.In(loc)
	}
	return data
}
