package mongo

import (
	"testing"
	"time"
)

func TestEntityArchived_IsArchived(t *testing.T) {
	tests := []struct {
		name     string
		archived IEntityArchived
		want     bool
	}{
		{
			name:     "not archived",
			archived: EntityArchived{},
			want:     false,
		},
		{
			name: "archived field",
			archived: EntityArchived{
				Archived: true,
			},
			want: true,
		},
		{
			name: "deleteTime field",
			archived: EntityArchived{
				DeleteTime: time.Now(),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.archived.IsArchived(); got != tt.want {
				t.Errorf("EntityArchived.IsArchived() = %v, want %v", got, tt.want)
			}
		})
	}
}
