package breaker

import (
	"time"
)

// settings configures CircuitBreaker:
//
// Name is the name of the CircuitBreaker.
//
// MaxRequests is the maximum number of requests allowed to pass through
// when the CircuitBreaker is half-open.
// If MaxRequests is 0, the CircuitBreaker allows only 1 request.
//
// Interval is the cyclic period of the closed state
// for the CircuitBreaker to clear the internal Counts.
// If Interval is less than or equal to 0, the CircuitBreaker doesn't clear internal Counts during the closed state.
//
// Timeout is the period of the open state,
// after which the state of the CircuitBreaker becomes half-open.
// If Timeout is less than or equal to 0, the timeout value of the CircuitBreaker is set to 60 seconds.
//
// ReadyToTrip is called with a copy of Counts whenever a request fails in the closed state.
// If ReadyToTrip returns true, the CircuitBreaker will be placed into the open state.
// If ReadyToTrip is nil, default ReadyToTrip is used.
// Default ReadyToTrip returns true when the number of consecutive failures is more than 5.
//
// OnStateChange is called whenever the state of the CircuitBreaker changes.
//
// IsSuccessful is called with the error returned from a request.
// If IsSuccessful returns true, the error is counted as a success.
// Otherwise the error is counted as a failure.
// If IsSuccessful is nil, default IsSuccessful is used, which returns false for all non-nil errors.
type settings struct {
	name          string
	maxRequests   uint32
	interval      time.Duration
	timeout       time.Duration
	readyToTrip   func(counts Counts) bool
	onStateChange func(name string, from State, to State)
	isSuccessful  func(err error) bool
}

func Settings() settings {
	return settings{
		name:         "",
		maxRequests:  1,
		interval:     time.Duration(0) * time.Second,
		timeout:      time.Duration(60) * time.Second,
		readyToTrip:  func(counts Counts) bool { return counts.ConsecutiveFailures > 5 },
		isSuccessful: func(err error) bool { return err == nil },
	}
}

func (s settings) WithName(name string) settings {
	s.name = name
	return s
}

func (s settings) WithMaxRequests(maxRequests uint32) settings {
	s.maxRequests = maxRequests
	return s
}

func (s settings) WithInterval(interval time.Duration) settings {
	if interval > 0 {
		s.interval = interval
	}
	return s
}

func (s settings) WithTimeout(timeout time.Duration) settings {
	if timeout > 0 {
		s.timeout = timeout
	}
	return s
}

func (s settings) WithReadyToTrip(readyToTrip func(counts Counts) bool) settings {
	s.readyToTrip = readyToTrip
	return s
}

func (s settings) WithOnStateChange(onStateChange func(name string, from State, to State)) settings {
	s.onStateChange = onStateChange
	return s
}

func (s settings) WithIsSuccessful(isSuccessful func(err error) bool) settings {
	s.isSuccessful = isSuccessful
	return s
}
