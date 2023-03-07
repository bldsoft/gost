package routing

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

type MultiAction struct {
	Actions []Action
}

func JoinActions(action ...Action) MultiAction {
	return MultiAction{
		Actions: action,
	}
}

func (a MultiAction) Apply(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, action := range a.Actions {
			h = action.Apply(h)
		}
		h.ServeHTTP(w, r)
	})
}

func (a MultiAction) MarshalBSON() ([]byte, error) {
	name, err := actionMarshaller.NameByValue(a)
	if err != nil {
		return nil, err
	}
	type tempAct struct {
		Actions []marshallingField[Action] `bson:"actions"`
	}
	return addTypeAndMashalBson(tempAct{makeMarshallingFields(actionMarshaller, a.Actions...)}, name)
}

func (a MultiAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Actions []marshallingField[Action] `json:"actions"`
	}{
		Actions: makeMarshallingFields(actionMarshaller, a.Actions...),
	})
}

func (a *MultiAction) UnmarshalJSON(b []byte) error {
	type outMultiAction struct {
		Actions []json.RawMessage `json:"actions"`
	}
	temp := &outMultiAction{}
	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	a.Actions = nil
	for _, actData := range temp.Actions {
		act, err := actionMarshaller.UnmarshalJSON(actData)
		if err != nil {
			return err
		}
		a.Actions = append(a.Actions, act)
	}
	return nil
}

func (a *MultiAction) UnmarshalBSON(b []byte) error {
	type outMultiAction struct {
		Actions []bson.Raw `bson:"actions"`
	}
	temp := &outMultiAction{}
	if err := bson.Unmarshal(b, &temp); err != nil {
		return err
	}

	a.Actions = nil
	for _, actData := range temp.Actions {
		act, err := actionMarshaller.UnmarshalBSON(actData)
		if err != nil {
			return err
		}
		a.Actions = append(a.Actions, act)
	}
	return nil
}
