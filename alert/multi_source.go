package alert

import (
	"context"
	"errors"
	"time"
)

type sourceWithNextTime struct {
	source   Source
	nextTime time.Time
}

type MultiSource struct {
	sourceToNextTime []sourceWithNextTime
}

func NewMultiSource(sources ...Source) *MultiSource {
	return new(MultiSource).AddSource(sources...)
}

func (s *MultiSource) AddSource(source ...Source) *MultiSource {
	now := time.Now()
	for _, source := range source {
		s.sourceToNextTime = append(s.sourceToNextTime, sourceWithNextTime{source, now})
	}
	return s
}

func (s *MultiSource) EvaluateAlerts(ctx context.Context) ([]Alert, time.Time, error) {
	if len(s.sourceToNextTime) == 0 {
		return nil, time.Time{}, errors.New("no sources added")
	}

	now := time.Now()
	minNextTime := s.sourceToNextTime[0].nextTime
	var res []Alert
	var errs error
	for i, sourceWithNextTime := range s.sourceToNextTime {
		source := sourceWithNextTime.source
		nextTime := sourceWithNextTime.nextTime

		if nextTime.After(now) {
			minNextTime = s.minTime(minNextTime, nextTime)
			continue
		}

		alerts, nextTime, err := source.EvaluateAlerts(ctx)
		if err != nil {
			errs = errors.Join(errs, err)
			nextTime = now.Add(5 * time.Minute)
		}
		s.sourceToNextTime[i].nextTime = nextTime
		minNextTime = s.minTime(minNextTime, nextTime)
		res = append(res, alerts...)
	}
	return res, minNextTime, errs
}

func (s *MultiSource) minTime(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
