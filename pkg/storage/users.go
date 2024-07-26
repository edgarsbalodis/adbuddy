package storage

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	UserID   int64              `bson:"user_id,omitempty"`
	ChatID   int64              `bson:"chat_id,omitempty"`
	Username string             `bson:"username"`
	IsActive bool               `bson:"is_active"`
}

func NewUser(userID int64, chatID int64, username string, isActive bool) *User {
	return &User{
		UserID:   userID,
		ChatID:   chatID,
		Username: username,
		IsActive: isActive,
	}
}

func (s *Storage) FindUser(userID int64) (User, error) {
	collection := s.GetCollection("users")

	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}
	// Retrieves the first matching document
	var existingUser User

	err := collection.FindOne(context.TODO(), filter).Decode(&existingUser)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return User{}, nil
		}

		if existingUser.UserID == userID {
			return User{}, err
		}
	}

	return existingUser, nil
}

func (s *Storage) SaveUser(user User) (User, error) {
	collection := s.GetCollection("users")

	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		return User{}, err
	}

	insertedID := result.InsertedID.(primitive.ObjectID)
	user.ID = insertedID

	return user, nil
}

// checks if user is allowed to communicate with bot
// it can be based on subscription/tokens whatever
func (s *Storage) IsUserValid(userID int64) bool {
	collection := s.GetCollection("users")

	filter := bson.D{
		primitive.E{Key: "user_id", Value: userID},
		primitive.E{Key: "is_active", Value: true},
	}

	// Retrieves the first matching document
	var existingUser User

	err := collection.FindOne(context.TODO(), filter).Decode(&existingUser)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false
		}
		return false
	}

	return true
}
