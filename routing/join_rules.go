package routing

import (
	"encoding/json"
	"net/http"

	"github.com/bldsoft/gost/log"
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

func (rl RuleList) Match(r *http.Request) (matched bool, err error) {
	if rl.Rule == nil {
		return true, nil
	}
	return rl.Rule.Match(r)
}

func (rl RuleList) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rl.Rule != nil {
			next = rl.Rule.Apply(next)
		}
		for _, rule := range rl.Rules {
			if rule == nil {
				continue
			}
			matched, err := rule.Match(r)
			switch {
			case matched:
				next = rule.Apply(next)
			case err != nil:
				log.FromContext(r.Context()).ErrorWithFields(log.Fields{"err": err}, "Routing: checking the rule condition for the request")
			}
		}
		next.ServeHTTP(w, r)
	})
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
