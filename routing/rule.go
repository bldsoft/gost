package routing

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson"
)

type Rule struct {
	Name string
	Condition
	Action
}

func NewRule(cond Condition, action Action) *Rule {
	return &Rule{
		Condition: cond,
		Action:    action,
	}
}

func (r *Rule) WithName(name string) *Rule {
	r.Name = name
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
		Name:      r.Name,
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
	r.Name = temp.Name
	r.Condition = temp.Condition.Value
	r.Action = temp.Action.Value
	return nil
}
