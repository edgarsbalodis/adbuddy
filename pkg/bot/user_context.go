package bot

import (
	"github.com/edgarsbalodis/adbuddy/pkg/storage"
)

type UserContext struct {
	ChatUser        *ChatUser
	CurrentQuestion int
	Questions       []storage.Question
	Answers         AnswersMap
}

type UserContextMap map[int64]*UserContext
type AnswersMap map[string]interface{}

func NewUserContext(ch *ChatUser, cq int, q []storage.Question) *UserContext {
	return &UserContext{
		ChatUser:        ch,
		CurrentQuestion: cq,
		Questions:       q,
		Answers:         make(AnswersMap),
	}
}
