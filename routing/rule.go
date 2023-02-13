package routing

type rule struct {
	Condition
	Action
}

func NewRule(cond Condition, action Action) Rule {
	return rule{
		Condition: cond,
		Action:    action,
	}
}
