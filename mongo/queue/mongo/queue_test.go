package mongo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/mongo"
)

var cfg mongo.Config

func init() {
	config.ReadConfig(&cfg, "")
}

func TestQueue(t *testing.T) {
	db := mongo.NewStorage(cfg)
	db.Connect()
	defer db.Disconnect(context.Background())

	queue := NewQueue[string](db, "test", Config{
		ProcessTimeout:    1 * time.Second,
		OrderedProcessing: true,
	})

	queue.Enqueue(context.Background(), "test1", "test2", "test3")

	items, err := queue.Dequeue(context.Background())
	if err != nil {
		t.Fatalf("failed to dequeue: %v", err)
	}

	fmt.Println(items)
	queue.AckItems(context.Background(), items)
	items, err = queue.Dequeue(context.Background())
	if err != nil {
		t.Fatalf("failed to dequeue: %v", err)
	}
	fmt.Println(items)
}
