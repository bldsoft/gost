package routing

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

type MultiCondition struct {
	Conditions []Condition
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

func (c MultiCondition) MarshalBSON() ([]byte, error) {
	name, err := conditionPolymorphMarshaller.NameByValue(c)
	if err != nil {
		return nil, err
	}
	type tempCond struct {
		Conditions []marshallingField[Condition] `bson:"conditions"`
	}
	return addTypeAndMashalBson(tempCond{Conditions: makeMarshallingFields(conditionPolymorphMarshaller, c.Conditions...)}, name)
}

func (c MultiCondition) MarshalJSON() ([]byte, error) {
	name, err := conditionPolymorphMarshaller.NameByValue(c)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Name       string                        `json:"type"`
		Conditions []marshallingField[Condition] `json:"conditions"`
	}{
		Name:       name,
		Conditions: makeMarshallingFields(conditionPolymorphMarshaller, c.Conditions...),
	})
}

func (rl *MultiCondition) UnmarshalJSON(b []byte) error {
	type outConditionList struct {
		Conditions []json.RawMessage `json:"conditions"`
	}
	temp := &outConditionList{}
	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	rl.Conditions = nil
	for _, condData := range temp.Conditions {
		cond, err := conditionPolymorphMarshaller.UnmarshalJSON(condData)
		if err != nil {
			return err
		}
		rl.Conditions = append(rl.Conditions, cond)
	}
	return nil
}

func (rl *MultiCondition) UnmarshalBSON(b []byte) error {
	type outConditionList struct {
		Conditions []bson.Raw `bson:"conditions"`
	}
	temp := &outConditionList{}
	if err := bson.Unmarshal(b, &temp); err != nil {
		return err
	}

	rl.Conditions = nil
	for _, condData := range temp.Conditions {
		cond, err := conditionPolymorphMarshaller.UnmarshalBSON(condData)
		if err != nil {
			return err
		}
		rl.Conditions = append(rl.Conditions, cond)
	}
	return nil
}
