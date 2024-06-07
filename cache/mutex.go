package cache

import (
	"bytes"
	"context"
	"crypto/rand"
	"time"

	"github.com/bldsoft/gost/log"
)

// DistrMutex - implementation of distributed lock
type DistrMutex struct {
	cache           IDistrCacheRepository
	lockKey         string
	uniqueID        []byte
	ticker          *time.Ticker
	quit            chan struct{}
	unlockTime      time.Duration
	TryLockInterval time.Duration
}

const defaultUnlockTime = time.Minute

type lockOwner int

const (
	nobody lockOwner = iota
	me
	notme
)

// NewDistrMutex creates an implementation of distributed lock.
// If the gouritine locks m and then finishes running without calling Unlock(), m unlocks after unlockTime.
func NewDistrMutex(cache IDistrCacheRepository, lockKey string, unlockTime time.Duration) *DistrMutex {
	uniqueID := make([]byte, 4)
	rand.Read(uniqueID)

	if unlockTime <= 0 {
		unlockTime = defaultUnlockTime
	}

	return &DistrMutex{
		cache:           cache,
		lockKey:         lockKey,
		uniqueID:        uniqueID,
		quit:            make(chan struct{}),
		unlockTime:      unlockTime,
		TryLockInterval: unlockTime,
	}
}

// Lock locks m. If the lock is already in use, the calling goroutine blocks until the mutex is available.
// LockKey record can be deleted according to LRU if Memcached fills up. In that case gourutine may pass through Lock too.
func (m *DistrMutex) Lock(ctx context.Context) {
	ticker := time.NewTicker(m.TryLockInterval)
	defer ticker.Stop()
	if m.TryLock() {
		return
	}

	for {
		select {
		case <-ticker.C:
			if m.TryLock() {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// TryLock tries to lock m. It returns true in case of success, false otherwise.
func (m *DistrMutex) TryLock() bool {
	err := m.cache.Add(m.lockKey, &Item{
		Value: m.uniqueID,
		TTL:   m.unlockTime,
	})
	if err != nil {
		log.DebugWithFields(log.Fields{"error": err}, "Failed to lock memcached mutex")
		return false
	}
	m.ticker = time.NewTicker(m.unlockTime / 2)
	go m.updateLock()
	return true
}

func (m *DistrMutex) getOwner() lockOwner {
	it, err := m.cache.Get(m.lockKey)
	if err != nil {
		return nobody
	}
	if bytes.Equal(it.Value, m.uniqueID) {
		return me
	}
	return notme
}

func (m *DistrMutex) updateLock() {
	for range m.ticker.C {
		lockOwner := m.getOwner()
		switch lockOwner {
		case me:
			err := m.cache.Set(m.lockKey, &Item{
				Value: m.uniqueID,
				TTL:   m.unlockTime,
			})
			if err != nil {
				log.ErrorfWithFields(log.Fields{"error": err}, "failed to update memcached lock %s", m.lockKey)
			}
		case notme:
			m.stop()
			return
		case nobody:
			if !m.TryLock() {
				m.stop()
			}
			return
		}
	}
}

func (m *DistrMutex) stop() {
	m.quit <- struct{}{}
	m.ticker.Stop()
}

// Quit is used to signal that lock doesn't belong to you anymore
func (m *DistrMutex) Quit() <-chan struct{} {
	return m.quit
}

// Unlock unlocks m if it belongs to you
func (m *DistrMutex) Unlock() {
	if m.ticker != nil {
		m.ticker.Stop()
	}
	if m.getOwner() == me {
		m.cache.Delete(m.lockKey)
	}
}
