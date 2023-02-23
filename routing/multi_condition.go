package routing

import "net/http"

type MultiCondition struct {
	Conditions []Condition `json:"conditions" bson:"conditions"`
}

func JoinConditions(conditions ...Condition) MultiCondition {
	return MultiCondition{
		Conditions: conditions,
	}
}

func (c MultiCondition) Match(r *http.Request) (matched bool, err error) {
	for _, condition := range c.Conditions {
		matched, err = condition.Match(r)
		if err != nil {
			return false, err
		}
		if !matched {
			return false, nil
		}
	}
	return true, nil
}
