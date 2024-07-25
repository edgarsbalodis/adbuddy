package bot

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
// var allowedIds = []int64{1268266402, 5948083859}
