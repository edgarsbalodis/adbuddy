package bot

import "golang.org/x/exp/slices"

// Contains all general and necessary information about telegram user
type ChatUser struct {
	UserID int64
	ChatID int64
}

func CreateChatUser(userID, chatID int64) *ChatUser {
	return &ChatUser{
		UserID: userID,
		ChatID: chatID,
	}
}

// When start I can also save this information in database

// validation function
var allowedIds = []int64{1268266402, 5948083859}

// checks if user is allowed to communicate with bot
// it can be based on subscription/tokens whatever
func isUserValid(userID int64) bool {
	return slices.Contains(allowedIds, userID)
}
