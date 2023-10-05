package questionnare

type Questionnare struct {
	Question string
	Options  OptionMap
	Key      string
	FollowUp bool
}

type QuestionnareList []*Questionnare
type OptionMap map[string]string

func NewQuestionnare(q string, opt OptionMap, k string) *Questionnare {
	return &Questionnare{
		Question: q,
		Options:  opt,
		Key:      k,
	}
}

// TODO

// return base questionnare with base question
