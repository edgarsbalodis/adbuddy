package bot

// Caintains all general and necessary information about telegram user
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
