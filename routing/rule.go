package routing

import "encoding/json"

type Rule struct {
	Condition
	Action
}

func NewRule(cond Condition, action Action) *Rule {
	return &Rule{
		Condition: cond,
		Action:    action,
	}
}

func (r *Rule) UnmarshalJSON(b []byte) error {
	type outRule struct {
		Condition marshallingField[Condition] `json:"condition"`
		Action    marshallingField[Action]    `json:"action"`
	}
	temp := &outRule{
		Condition: marshallingField[Condition]{polymorphMarshaller: conditionPolymorphMarshaller},
		Action:    marshallingField[Action]{polymorphMarshaller: actionMarshaller},
	}
	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}
	r.Condition = temp.Condition.Value
	r.Action = temp.Action.Value
	return nil
}

func (r Rule) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Condition marshallingField[Condition] `json:"condition"`
		Action    marshallingField[Action]    `json:"action"`
	}{
		Condition: marshallingField[Condition]{r.Condition, conditionPolymorphMarshaller},
		Action:    marshallingField[Action]{r.Action, actionMarshaller},
	})
}
