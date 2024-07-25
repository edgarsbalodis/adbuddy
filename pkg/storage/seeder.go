package storage

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TODO:
//	[ ] create json files for questions
//	[ ] create functionality for migrating data with json files

func (s *Storage) Seed() {
	s.QuestionSeeder()
	s.FilterSeeder()
}

func (s *Storage) QuestionSeeder() {
	// check count of questions, if they already exist in database
	coll := s.GetCollection("questions")
	opts := options.Count().SetHint("_id_")
	count, _ := coll.CountDocuments(context.TODO(), bson.D{}, opts)

	if count > 0 {
		return
	}
	// create index
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "options.condition", Value: 1}},
	}

	name, err := coll.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Fatal("couldn't create index")
	}

	fmt.Printf("Index created: %v\n", name)

	// create q's
	type questions []interface{}
	q := questions{
		&Question{
			Text:     "What are you looking for?",
			Main:     true,
			Key:      "Type",
			FollowUp: false,
		},
		&Question{
			Text:     "Do you want to buy or rent?",
			Key:      "TransactionType",
			FollowUp: false,
			Options: []Option{
				{Text: "Buy", Value: "buy"},
				{Text: "Rent", Value: "rent"},
			},
		},
		&Question{
			Text:     "Write in chat, what's your max price? For example: 300000",
			Key:      "Price",
			FollowUp: false,
		},
		&Question{
			Text:     "Write in chat, what's min number of rooms? For example: 4",
			Key:      "NumOfRooms",
			FollowUp: false,
		},
		&Question{
			Text:     "Choose region/city",
			Key:      "Region",
			FollowUp: false,
			Options: []Option{
				{Text: "Rīgas raj.", Value: "riga-region"},
				{Text: "Rīga", Value: "riga"},
				{Text: "Jūrmala", Value: "jurmala"},
			},
		},
		&Question{
			Text:     "Choose subregion",
			Key:      "Subregion",
			FollowUp: true,
			Options: []Option{
				{Text: "Mārupes pag.", Value: "marupes-pag", Condition: "riga-region"},
				{Text: "Babītes pag.", Value: "babites-pag", Condition: "riga-region"},
				{Text: "Centre", Value: "centre", Condition: "riga"},
				{Text: "Āgenskalns", Value: "agenskalns", Condition: "riga"},
			},
		},
	}

	res, err := coll.InsertMany(context.TODO(), q)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Question documents inserted: %v\n", len(res.InsertedIDs))
}

func (s *Storage) FilterSeeder() {
	// check count of questions, if they already exist in database
	coll := s.GetCollection("filters")
	opts := options.Count().SetHint("_id_")
	count, _ := coll.CountDocuments(context.TODO(), bson.D{}, opts)

	if count > 0 {
		return
	}

	qcoll := s.GetCollection("questions")

	cur, err := qcoll.Find(context.TODO(), bson.M{
		"main": bson.M{"$ne": true},
	})
	if err != nil {
		log.Fatal(err)
	}
	var results []primitive.ObjectID
	for cur.Next(context.TODO()) {
		//Create a value into which the single document can be decoded
		var elem Question
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, elem.ID)
	}
	fmt.Print(results)

	type filters []interface{}

	f := filters{
		&Filter{
			Name:      "Residence",
			Value:     "residence",
			Questions: results,
		},
		&Filter{
			Name:      "Flat",
			Value:     "flat",
			Questions: results,
		},
	}

	res, err := coll.InsertMany(context.TODO(), f)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Filters documents inserted: %v\n", len(res.InsertedIDs))
}
