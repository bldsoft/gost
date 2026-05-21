//go:build integration_test

package mongolock

import (
	"context"
	"sort"

	"testing"
	"time"
)

func TestPurge(t *testing.T) {
	// setup and teardown are defined in lock_test.go
	collection := setup(t)
	defer teardown(t, collection)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	client := NewClient(collection)

	// Create some locks.
	err := client.XLock(ctx, "resource1", "aaaa", LockDetails{})
	if err != nil {
		t.Error(err)
	}
	err = client.XLock(ctx, "resource2", "bbbb", LockDetails{TTL: 1})
	if err != nil {
		t.Error(err)
	}
	err = client.SLock(ctx, "resource3", "cccc", LockDetails{TTL: 1}, -1)
	if err != nil {
		t.Error(err)
	}

	// Sleep for a second to let TTLs expire
	time.Sleep(time.Duration(1500) * time.Millisecond)

	// Purge the locks.
	purger := NewPurger(client)
	purged, err := purger.Purge(ctx)
	if err != nil {
		t.Error(err)
	}

	if len(purged) != 2 {
		t.Errorf("%d locks purged, expected %d", len(purged), 2)
	}

	var purgedSorted LockStatusesByCreatedAtDesc
	purgedSorted = purged
	sort.Sort(purgedSorted)
	if purged[0].Resource != "resource3" {
		t.Errorf("purged[0].Resource = %s, expected %s", purged[0].Resource, "resource3")
	}
	if purged[1].Resource != "resource2" {
		t.Errorf("purged[1].Resource = %s, expected %s", purged[1].Resource, "resource2")
	}
}

func TestPurgeSameLockIdDiffTTLs(t *testing.T) {
	// setup and teardown are defined in lock_test.go
	collection := setup(t)
	defer teardown(t, collection)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	client := NewClient(collection)

	// Create some locks with different TTLs, all owned by the same lockId.
	err := client.XLock(ctx, "resource1", "aaaa", LockDetails{}) // no TTL
	if err != nil {
		t.Error(err)
	}
	err = client.XLock(ctx, "resource2", "aaaa", LockDetails{TTL: 30})
	if err != nil {
		t.Error(err)
	}
	err = client.SLock(ctx, "resource3", "aaaa", LockDetails{TTL: 1}, -1)
	if err != nil {
		t.Error(err)
	}

	// Sleep for a second to let some TTLs expire
	time.Sleep(time.Duration(1500) * time.Millisecond)

	// Purge the locks.
	purger := NewPurger(client)
	purged, err := purger.Purge(ctx)
	if err != nil {
		t.Error(err)
	}

	if len(purged) != 3 {
		t.Errorf("%d locks purged, expected %d", len(purged), 3)
	}

	var purgedSorted LockStatusesByCreatedAtDesc
	purgedSorted = purged
	sort.Sort(purgedSorted)
	if purged[0].Resource != "resource3" {
		t.Errorf("purged[0].Resource = %s, expected %s", purged[0].Resource, "resource3")
	}
	if purged[1].Resource != "resource2" {
		t.Errorf("purged[1].Resource = %s, expected %s", purged[1].Resource, "resource2")
	}
	if purged[2].Resource != "resource1" {
		t.Errorf("purged[2].Resource = %s, expected %s", purged[2].Resource, "resource1")
	}
}
