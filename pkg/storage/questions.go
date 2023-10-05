package storage

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Questions struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`                // id	// id
	RelatedQuestionID primitive.ObjectID `bson:"related_questionID,omitempty"` // "" // <- id
	Text              string             `bson:"text"`                         // Choose Region // Choose subregion
	Main              bool               `bson:"main"`                         // false // false
	Key               string             `bson:"key"`                          // Region // Subregion
	FollowUp          bool               `bson:"follow_up"`                    // false // true
	Options           []Option           `bson:"options,omitempty"`
}

type Option struct {
	Key   string             `bson:"key"`                  // riga-region // marupes-pag
	Value string             `bson:"value"`                // riga region //marupes pag.
	ID    primitive.ObjectID `bson:"related_id,omitempty"` // sub regions //
}

// func (c *mongo.Client)
