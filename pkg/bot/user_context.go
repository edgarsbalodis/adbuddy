package bot

import "github.com/edgarsbalodis/adbuddy/pkg/questionnare"

type UserContext struct {
	ChatUser        *ChatUser
	CurrentQuestion int
	Questionnares   questionnare.QuestionnareList
	Answers         AnswersMap
}

type UserContextMap map[int64]*UserContext

// type AnswersMap map[string]string
type AnswersMap map[string]interface{}

func NewUserContext(ch *ChatUser, cq int, q questionnare.QuestionnareList) *UserContext {
	return &UserContext{
		ChatUser:        ch,
		CurrentQuestion: cq,
		Questionnares:   q,
		Answers:         make(AnswersMap),
	}
}
