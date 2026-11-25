package service

import (
	"context"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
)

func NewPurger(notes NoteRepository, interval, after time.Duration) server.AsyncRunner {
	return server.NewContextAsyncRunner(func(ctx context.Context) error {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if err := notes.PurgeDeletedBefore(ctx, time.Now().Add(-after)); err != nil {
					log.ErrorWithFields(log.Fields{"error": err}, "Failed to purge soft-deleted notes")
				}
			}
		}
	})
}
