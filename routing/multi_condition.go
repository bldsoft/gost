package routing

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

type MultiCondition struct {
	MatchAny   bool
	Conditions []Condition
}

func JoinConditions(matchAll bool, conditions ...Condition) MultiCondition {
	return MultiCondition{
		MatchAny:   !matchAll,
		Conditions: conditions,
	}
}

func (c MultiCondition) MatchAll() bool {
	return !c.MatchAny
}

func (c MultiCondition) IncomingMatch(w http.ResponseWriter, r *http.Request) (matched bool, outgoingMatch outgoingMatchFunc, err error) {
	var outgoingMatches []outgoingMatchFunc
	for _, condition := range c.Conditions {
		matched, outgoingMatch, err := condition.IncomingMatch(w, r)
		if err != nil {
			return false, nil, err
		}

		if outgoingMatch != nil {
			outgoingMatches = append(outgoingMatches, outgoingMatch)
			continue
		}

		if c.MatchAny && matched {
			return matched, nil, nil
		}

		if c.MatchAll() && !matched {
			return matched, nil, nil
		}
	}

	if len(outgoingMatches) == 0 {
		return c.MatchAll(), nil, nil
	}

	return false, func(w http.ResponseWriter, r *http.Request) (matched bool, err error) {
		for _, match := range outgoingMatches {
			matched, err := match(w, r)
			if err != nil {
				return false, err
			}

			if c.MatchAny && matched {
				return matched, nil
			}

			if c.MatchAll() && !matched {
				return matched, nil
			}
		}
		return c.MatchAll(), nil
	}, nil
}

func (c MultiCondition) MarshalBSON() ([]byte, error) {
	name, err := conditionPolymorphMarshaller.NameByValue(c)
	if err != nil {
		return nil, err
	}
	type tempCond struct {
		MatchAny   bool                          `bson:"matchAny"`
		Conditions []marshallingField[Condition] `bson:"conditions"`
	}
	return addTypeAndMashalBson(tempCond{
		MatchAny:   c.MatchAny,
		Conditions: makeMarshallingFields(conditionPolymorphMarshaller, c.Conditions...),
	}, name)
}

func (c MultiCondition) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		MatchAny   bool                          `json:"matchAny"`
		Conditions []marshallingField[Condition] `json:"conditions"`
	}{
		MatchAny:   c.MatchAny,
		Conditions: makeMarshallingFields(conditionPolymorphMarshaller, c.Conditions...),
	})
}

func (rl *MultiCondition) UnmarshalJSON(b []byte) error {
	type outConditionList struct {
		MatchAny   bool              `json:"matchAny"`
		Conditions []json.RawMessage `json:"conditions"`
	}
	temp := &outConditionList{}
	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	rl.Conditions = nil
	rl.MatchAny = temp.MatchAny
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
		MatchAny   bool       `bson:"matchAny"`
		Conditions []bson.Raw `bson:"conditions"`
	}
	temp := &outConditionList{}
	if err := bson.Unmarshal(b, &temp); err != nil {
		return err
	}

	rl.Conditions = nil
	rl.MatchAny = temp.MatchAny
	for _, condData := range temp.Conditions {
		cond, err := conditionPolymorphMarshaller.UnmarshalBSON(condData)
		if err != nil {
			return err
		}
		rl.Conditions = append(rl.Conditions, cond)
	}
	return nil
}
