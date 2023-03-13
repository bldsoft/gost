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

func (a MultiAction) DoBeforeHandle(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error) {
	var err error
	for _, action := range a.Actions {
		w, r, err = action.DoBeforeHandle(w, r)
		if err != nil {
			return nil, nil, err
		}
	}
	return w, r, nil
}

func (a MultiAction) DoAfterHandle(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error) {
	var err error
	for _, action := range a.Actions {
		w, r, err = action.DoAfterHandle(w, r)
		if err != nil {
			return nil, nil, err
		}
	}
	return w, r, nil
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
