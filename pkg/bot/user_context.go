package bot

import "github.com/edgarsbalodis/adbuddy/pkg/questionnare"

type UserContext struct {
	ChatUser        *ChatUser
	CurrentQuestion int
	Questionnares   questionnare.QuestionnareList
	Answers         AnswersMap // answers will be saved in db
}

type UserContextMap map[int64]*UserContext
type AnswersMap map[string]string

func NewUserContext(ch *ChatUser, cq int, q questionnare.QuestionnareList) *UserContext {
	return &UserContext{
		ChatUser:        ch,
		CurrentQuestion: cq,
		Questionnares:   q,
		Answers:         make(AnswersMap),
	}
}
