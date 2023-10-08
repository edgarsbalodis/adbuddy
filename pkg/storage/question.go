package storage

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Question struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`                // id	// id
	ParentQuestionID primitive.ObjectID `bson:"parent_question_id,omitempty"` // "" // <- id
	Text             string             `bson:"text"`                         // Choose Region // Choose subregion
	Main             bool               `bson:"main,omitempty"`               // false // false
	Key              string             `bson:"key"`                          // Region // Subregion
	FollowUp         bool               `bson:"follow_up"`                    // false // true
	Options          []Option           `bson:"options,omitempty"`
}

type Option struct {
	Text      string `bson:"text"`  // riga-region // marupes-pag
	Value     string `bson:"value"` // riga region //marupes pag.
	Condition string `bsob:"condition,omitempty"`
}

func (s *Storage) GetMainQuestion() Question {
	coll := s.GetCollection("questions")

	var q Question
	err := coll.FindOne(context.TODO(), bson.M{"main": true}).Decode(&q)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Document not found
			fmt.Println("Document not found.")
		} else {
			// Another type of error occurred
			log.Fatal(err)
		}
		return Question{}
	}
	return q
}

// Based on filter type returns all questions
func (s *Storage) GetFiltersQuestions(filter string) []Question {
	coll := s.GetCollection("filters")

	var f Filter
	err := coll.FindOne(context.TODO(), bson.M{"value": filter}).Decode(&f)
	if err != nil {
		log.Fatal(err)
	}

	collq := s.GetCollection("questions")
	qf := bson.M{"_id": bson.M{"$in": f.Questions}}

	cursor, err := collq.Find(context.TODO(), qf)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	var questions []Question
	if err = cursor.All(context.TODO(), &questions); err != nil {
		log.Fatal(err)
	}

	return questions
}
