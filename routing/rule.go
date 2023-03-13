package routing

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

type Rule struct {
	name string
	Condition
	Action
}

func NewRule(cond Condition, action Action) *Rule {
	return &Rule{
		Condition: cond,
		Action:    action,
	}
}

func (rule Rule) Name() string {
	return rule.name
}

func (rule Rule) IncomingMatch(w http.ResponseWriter, r *http.Request) (matched bool, outgoingMatch outgoingMatchFunc, err error) {
	if rule.Condition != nil {
		return rule.Condition.IncomingMatch(w, r)
	}
	return true, nil, nil
}

func (rule Rule) DoBeforeHandle(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error) {
	if rule.Action != nil {
		return rule.Action.DoBeforeHandle(w, r)
	}
	return w, r, nil
}

func (rule Rule) DoAfterHandle(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error) {
	if rule.Action != nil {
		return rule.Action.DoAfterHandle(w, r)
	}
	return w, r, nil
}

func (r *Rule) WithName(name string) *Rule {
	r.name = name
	return r
}

func (r Rule) MarshalJSON() ([]byte, error) {
	return r.marshalHelper(json.Marshal)
}

func (r Rule) MarshalBSON() ([]byte, error) {
	return r.marshalHelper(bson.Marshal)
}

func (r *Rule) UnmarshalJSON(b []byte) error {
	return r.unmarshalHelper(b, json.Unmarshal)
}

func (r *Rule) UnmarshalBSON(b []byte) error {
	return r.unmarshalHelper(b, bson.Unmarshal)
}

func (r Rule) marshalHelper(marshalFunc func(val interface{}) ([]byte, error)) ([]byte, error) {
	return marshalFunc(struct {
		Name      string                      `json:"name,omitempty" bson:"name,omitempty"`
		Condition marshallingField[Condition] `json:"condition" bson:"condition"`
		Action    marshallingField[Action]    `json:"action" bson:"action"`
	}{
		Name:      r.name,
		Condition: marshallingField[Condition]{r.Condition, conditionPolymorphMarshaller},
		Action:    marshallingField[Action]{r.Action, actionMarshaller},
	})
}

func (r *Rule) unmarshalHelper(b []byte, unmarshalFunc func(data []byte, v any) error) error {
	type outRule struct {
		Name      string                      `json:"name,omitempty" bson:"name,omitempty"`
		Condition marshallingField[Condition] `json:"condition" bson:"condition"`
		Action    marshallingField[Action]    `json:"action" bson:"action"`
	}
	temp := &outRule{
		Condition: marshallingField[Condition]{polymorphMarshaller: conditionPolymorphMarshaller},
		Action:    marshallingField[Action]{polymorphMarshaller: actionMarshaller},
	}
	if err := unmarshalFunc(b, &temp); err != nil {
		return err
	}
	r.name = temp.Name
	r.Condition = temp.Condition.Value
	r.Action = temp.Action.Value
	return nil
}
