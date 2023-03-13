package routing

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

type RuleList struct {
	Rule  IRule
	Rules []IRule
}

func JoinRules(rules ...IRule) *RuleList {
	return &RuleList{
		Rules: rules,
	}
}

func (rl RuleList) Name() string {
	if rl.Rule != nil {
		rl.Rule.Name()
	}
	return ""
}

func (rl RuleList) IncomingMatch(w http.ResponseWriter, r *http.Request) (matched bool, outgoingMatch outgoingMatchFunc, err error) {
	if rl.Rule != nil {
		return rl.Rule.IncomingMatch(w, r)
	}
	return true, nil, nil
}

func (rl RuleList) DoBeforeHandle(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error) {
	if rl.Rule != nil {
		return rl.Rule.DoBeforeHandle(w, r)
	}
	return w, r, nil
}
func (rl RuleList) DoAfterHandle(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error) {
	if rl.Rule != nil {
		return rl.Rule.DoAfterHandle(w, r)
	}
	return w, r, nil
}

func (rl RuleList) MarshalBSON() ([]byte, error) {
	name, err := ruleMarshaller.NameByValue(&rl)
	if err != nil {
		return nil, err
	}
	type tempRuleList struct {
		Rule  *marshallingField[IRule]  `json:"rule,omitempty"`
		Rules []marshallingField[IRule] `json:"rules"`
	}
	return addTypeAndMashalBson(tempRuleList{
		Rule:  newMarshallingField(ruleMarshaller, rl.Rule),
		Rules: makeMarshallingFields(ruleMarshaller, rl.Rules...),
	}, name)
}

func (rl RuleList) MarshalJSON() ([]byte, error) {
	name, err := ruleMarshaller.NameByValue(&rl)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Name  string                    `json:"type"`
		Rule  *marshallingField[IRule]  `json:"rule,omitempty"`
		Rules []marshallingField[IRule] `json:"rules"`
	}{
		Name:  name,
		Rule:  newMarshallingField(ruleMarshaller, rl.Rule),
		Rules: makeMarshallingFields(ruleMarshaller, rl.Rules...),
	})
}

func (rl *RuleList) UnmarshalJSON(b []byte) error {
	type outRuleList struct {
		Rule  marshallingField[IRule] `json:"rule"`
		Rules []json.RawMessage       `json:"rules"`
	}
	temp := &outRuleList{Rule: marshallingField[IRule]{polymorphMarshaller: ruleMarshaller}}
	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	rl.Rules = nil
	rl.Rule = temp.Rule.Value
	for _, ruleData := range temp.Rules {
		rule, err := ruleMarshaller.UnmarshalJSON(ruleData)
		if err != nil {
			return err
		}
		rl.Rules = append(rl.Rules, rule)
	}
	return nil
}

func (rl *RuleList) UnmarshalBSON(b []byte) (err error) {
	type outRuleList struct {
		Rule  bson.Raw   `bson:"rule"`
		Rules []bson.Raw `bson:"rules"`
	}
	temp := &outRuleList{}
	if err := bson.Unmarshal(b, &temp); err != nil {
		return err
	}

	rl.Rules = nil
	if len(temp.Rule) > 0 {
		if rl.Rule, err = ruleMarshaller.UnmarshalBSON(temp.Rule); err != nil {
			return err
		}
	}
	for _, ruleData := range temp.Rules {
		rule, err := ruleMarshaller.UnmarshalBSON(ruleData)
		if err != nil {
			return err
		}
		rl.Rules = append(rl.Rules, rule)
	}
	return nil
}
