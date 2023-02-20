package routing

import (
	"encoding/json"
	"net/http"

	"github.com/bldsoft/gost/log"
	"go.mongodb.org/mongo-driver/bson"
)

type RuleList struct {
	Condition
	Rules []IRule `json:"rules,omitempty" bson:"rules,omtempty"`
}

func JoinRules(rules ...IRule) RuleList {
	return RuleList{
		Rules: rules,
	}
}

func (rl RuleList) Match(r *http.Request) (matched bool, err error) {
	if rl.Condition == nil {
		return true, nil
	}
	return rl.Condition.Match(r)
}

func (rl RuleList) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func (rl *RuleList) UnmarshalJSON(b []byte) error {
	type outRuleList struct {
		Rules []json.RawMessage `json:"rules"`
	}
	temp := &outRuleList{}
	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	rl.Rules = nil
	for _, ruleData := range temp.Rules {
		rule, err := ruleMarshaller.UnmarshalJSON(ruleData)
		if err != nil {
			return err
		}
		rl.Rules = append(rl.Rules, rule)
	}
	return nil
}

func (rl *RuleList) UnmarshalBSON(b []byte) error {
	type outRuleList struct {
		Rules []bson.Raw `json:"rules"`
	}
	temp := &outRuleList{}
	if err := bson.Unmarshal(b, &temp); err != nil {
		return err
	}

	rl.Rules = nil
	for _, ruleData := range temp.Rules {
		rule, err := ruleMarshaller.UnmarshalBSON(ruleData)
		if err != nil {
			return err
		}
		rl.Rules = append(rl.Rules, rule)
	}
	return nil
}
