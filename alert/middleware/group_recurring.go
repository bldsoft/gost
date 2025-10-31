package middleware

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/bldsoft/gost/alert"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils"
)

type Group struct {
	ID     string
	ExpAt  time.Time
	Alerts []alert.Alert
}

type GroupFilter struct {
	IDs         []string
	ExpNotAfter time.Time
}

type GroupRepository interface {
	CreateGroup(ctx context.Context, group *Group) error
	UpdateGroup(ctx context.Context, group *Group) error
	FindGroups(ctx context.Context, filter GroupFilter) ([]*Group, error)
	Delete(ctx context.Context, filter GroupFilter) error
}

type GroupMiddleware struct {
	groupID                 string
	groupRep                GroupRepository
	groupPeriod             time.Duration
	checkExpiredGroupPeriod time.Duration

	alertMergeFunc func(alerts ...alert.Alert) alert.Alert
}

// use after deduplication middleware only, not concurrent safe
func NewGroupMiddleware(groupID string, groupRep GroupRepository, groupPeriod time.Duration) *GroupMiddleware {
	return &GroupMiddleware{
		groupID:     groupID,
		groupRep:    groupRep,
		groupPeriod: groupPeriod,
	}
}

func (m *GroupMiddleware) WithCheckExpiredGroupPeriod(checkExpiredGroupPeriod time.Duration) *GroupMiddleware {
	m.checkExpiredGroupPeriod = checkExpiredGroupPeriod
	return m
}

func (m *GroupMiddleware) Middleware() (_ alert.Middleware, close func()) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	return func(next alert.Handler) alert.Handler {
			wg.Add(1)
			go func() {
				defer wg.Done()
				m.run(ctx, next)
			}()

			return alert.HandlerFunc(func(ctx context.Context, alerts ...alert.Alert) {
				logger := log.FromContext(ctx).WithFields(log.Fields{"component": "alerts group middleware"})
				passed := make([]alert.Alert, 0, len(alerts))
				for _, a := range alerts {
					passedAlerts, err := m.processAlert(ctx, a)
					if err != nil {
						logger.ErrorWithFields(log.Fields{"err": err}, "failed to process alert")
						continue
					}
					passed = append(passed, passedAlerts...)
				}

				if len(passed) == 0 {
					return
				}
				next.Handle(ctx, passed...)
			})
		}, func() {
			cancel()
			wg.Wait()
		}
}

func (m *GroupMiddleware) processAlert(ctx context.Context, a alert.Alert) (passed []alert.Alert, err error) {
	var group *Group
	defer func() {
		if err != nil {
			return
		}
		if pass := m.passAlert(ctx, group); pass {
			passed = []alert.Alert{a}
		}
	}()

	now := time.Now()
	groups, err := m.groupRep.FindGroups(ctx, GroupFilter{
		IDs:         []string{m.groupID},
		ExpNotAfter: now.Add(m.groupPeriod),
	})
	if err != nil && !errors.Is(err, utils.ErrObjectNotFound) {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	if len(groups) == 0 {
		log.FromContext(ctx).DebugWithFields(log.Fields{"alert": a}, "no group found, creating new one")
		group = &Group{
			ID:     m.groupID,
			ExpAt:  now.Add(m.groupPeriod),
			Alerts: []alert.Alert{a},
		}

		if err = m.groupRep.CreateGroup(ctx, group); err != nil {
			return nil, fmt.Errorf("failed to create group: %w", err)
		}
		return nil, nil
	}

	group = groups[0]
	group.Alerts = append(group.Alerts, a)
	err = m.groupRep.UpdateGroup(ctx, group)
	if err == nil {
		return nil, nil
	}
	if !errors.Is(err, utils.ErrObjectNotFound) {
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	// deleted by run goroutine
	log.FromContext(ctx).DebugWithFields(log.Fields{"alert": a}, "group was deleted, creating new one")
	group.Alerts = []alert.Alert{a}
	group.ExpAt = a.From.Add(m.groupPeriod)
	if err = m.groupRep.CreateGroup(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}
	return nil, nil
}

func (m *GroupMiddleware) passAlert(ctx context.Context, group *Group) bool {
	// currentAlert := group.Alerts[len(group.Alerts)-1]

	if len(group.Alerts) == 0 { // just in case
		return false
	}

	if len(group.Alerts) == 1 {
		return true
	}

	if len(group.Alerts) == 2 && group.Alerts[0].To.IsZero() && !group.Alerts[1].To.IsZero() {
		log.FromContext(ctx).DebugWithFields(log.Fields{"start": group.Alerts[0], "finish": group.Alerts[1]}, "finish first alert in group")
		return true
	}
	return false
}

func (m *GroupMiddleware) run(ctx context.Context, next alert.Handler) error {
	const minCheckPeriod, maxCheckPeriod = 1 * time.Minute, 5 * time.Minute
	checkGroupPeriod := min(max(m.groupPeriod/2, minCheckPeriod), maxCheckPeriod)
	ticker := time.NewTicker(cmp.Or(m.checkExpiredGroupPeriod, checkGroupPeriod))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			m.handleGroups(ctx, time.Now().Add(m.groupPeriod), next)
			return nil
		case now := <-ticker.C:
			m.handleGroups(ctx, now, next)
		}
	}
}

func (m *GroupMiddleware) handleGroups(ctx context.Context, now time.Time, next alert.Handler) {
	groups, err := m.groupRep.FindGroups(ctx, GroupFilter{
		IDs:         []string{m.groupID},
		ExpNotAfter: now,
	})
	if err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "failed to get alert groups")
		return
	}
	if len(groups) == 0 {
		return
	}

	alerts := make([]alert.Alert, 0, len(groups))
	for _, group := range groups {
		if alreadyPassed := m.passAlert(ctx, group); alreadyPassed { // to avoid duplicated notifications
			continue
		}
		alerts = append(alerts, m.mergeAlerts(group.Alerts...))
	}
	if len(alerts) > 0 {
		next.Handle(ctx, alerts...)
	}

	m.groupRep.Delete(ctx, GroupFilter{
		IDs:         []string{m.groupID},
		ExpNotAfter: now,
	})
	log.FromContext(ctx).DebugOrErrorWithFields(err, log.Fields{"group": groups, "now": now}, "delete alert groups")
}

func (m *GroupMiddleware) WithAlertMergeFunc(alertMergeFunc func(alerts ...alert.Alert) alert.Alert) *GroupMiddleware {
	m.alertMergeFunc = alertMergeFunc
	return m
}

func (m *GroupMiddleware) mergeAlerts(alerts ...alert.Alert) alert.Alert {
	if m.alertMergeFunc != nil {
		return m.alertMergeFunc(alerts...)
	}
	return m.defaultAlertMergeFunc(alerts...)
}

func (m *GroupMiddleware) defaultAlertMergeFunc(alerts ...alert.Alert) alert.Alert {
	lastAlert := alerts[len(alerts)-1]
	res := alert.Alert{
		SourceID:  alerts[0].SourceID,
		Severity:  alerts[0].Severity,
		From:      alerts[0].From,
		To:        lastAlert.To,
		Receivers: lastAlert.Receivers, // the most recent receivers
		MetaData:  lastAlert.MetaData,
	}

	count := 0
	for _, alert := range alerts {
		if !alert.To.IsZero() {
			count++
		}
	}
	if res.To.IsZero() {
		count++
	}
	return res.AddMetaData("count", strconv.Itoa(count))
}
