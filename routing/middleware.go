package routing

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/utils"
)

type outgoingRule struct {
	match  outgoingMatchFunc
	action actionFunc
}

func (o outgoingRule) Match(w http.ResponseWriter, r *http.Request) (matched bool, err error) {
	if o.match != nil {
		return o.match(w, r)
	}
	return true, nil
}

type applyIncomingRulesRes struct {
	w             http.ResponseWriter
	r             *http.Request
	outgoingRules []outgoingRule
}

func applyIncomingRules(rules []IRule, res *applyIncomingRulesRes) error {
	for _, rule := range rules {
		matched, outgoingMatch, err := rule.IncomingMatch(res.w, res.r)
		if err != nil {
			return fmt.Errorf("incoming routing match failed for rule %s: %w", rule.Name(), err)
		}

		if outgoingMatch != nil {
			res.outgoingRules = append(res.outgoingRules, outgoingRule{
				match:  outgoingMatch,
				action: rule.DoAfterHandle,
			})
			continue
		}

		if matched {
			res.w, res.r, err = rule.DoBeforeHandle(res.w, res.r)
			if err != nil {
				return err
			}
			if ruleList, ok := rule.(*RuleList); ok {
				if err := applyIncomingRules(ruleList.Rules, res); err != nil {
					return err
				}
			}
			res.outgoingRules = append(res.outgoingRules, outgoingRule{
				action: rule.DoAfterHandle,
			})
		}
	}
	return nil
}

func Routing(rules ...IRule) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// incoming
			ww := utils.WrapResponseWriter(w)
			defer ww.Flush()
			applyRulesRes := applyIncomingRulesRes{
				w:             ww,
				r:             r,
				outgoingRules: make([]outgoingRule, 0, len(rules)),
			}
			if err := applyIncomingRules(rules, &applyRulesRes); err != nil {
				if !errors.Is(err, ErrStopHandling) {
					controller.ResponseError(applyRulesRes.w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
				return
			}

			// handle
			next.ServeHTTP(applyRulesRes.w, applyRulesRes.r)

			// outgoing
			for _, rule := range applyRulesRes.outgoingRules {
				matched, err := rule.Match(applyRulesRes.w, applyRulesRes.r)
				if err != nil {
					controller.ResponseError(applyRulesRes.w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}

				if !matched {
					continue
				}

				applyRulesRes.w, applyRulesRes.r, err = rule.action(applyRulesRes.w, applyRulesRes.r)
				if err != nil {
					if !errors.Is(err, ErrStopHandling) {
						controller.ResponseError(applyRulesRes.w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					}
					return
				}
			}
		})
	}
}
