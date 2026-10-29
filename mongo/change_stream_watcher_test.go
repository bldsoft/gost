package mongo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var w changeStreamWatcher

func TestChangeEventDocument(t *testing.T) {
	doc := bson.D{{Key: "_id", Value: "id"}, {Key: "field", Value: "value"}}
	key := bson.D{{Key: "_id", Value: "id"}}

	tests := []struct {
		name          string
		operationType string
		event         bson.D
		want          bson.D
		wantOK        bool
	}{
		{
			name:          "update with full document",
			operationType: changeStreamUpdateOp,
			event:         bson.D{{Key: "fullDocument", Value: doc}},
			want:          doc,
			wantOK:        true,
		},
		{
			name:          "update of already deleted document",
			operationType: changeStreamUpdateOp,
			event:         bson.D{{Key: "fullDocument", Value: nil}, {Key: "documentKey", Value: key}},
		},
		{
			name:          "replace without full document",
			operationType: changeStreamReplaceOp,
			event:         bson.D{{Key: "documentKey", Value: key}},
		},
		{
			name:          "delete with pre-image",
			operationType: changeStreamDeleteOp,
			event:         bson.D{{Key: "fullDocumentBeforeChange", Value: doc}, {Key: "documentKey", Value: key}},
			want:          doc,
			wantOK:        true,
		},
		{
			name:          "delete with null pre-image",
			operationType: changeStreamDeleteOp,
			event:         bson.D{{Key: "fullDocumentBeforeChange", Value: nil}, {Key: "documentKey", Value: key}},
			want:          key,
			wantOK:        true,
		},
		{
			name:          "delete without pre-image field",
			operationType: changeStreamDeleteOp,
			event:         bson.D{{Key: "documentKey", Value: key}},
			want:          key,
			wantOK:        true,
		},
		{
			name:          "delete without document key",
			operationType: changeStreamDeleteOp,
			event:         bson.D{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := bson.Marshal(tt.event)
			require.NoError(t, err)

			got, ok := w.changeEventDocument(event, tt.operationType)

			assert.Equal(t, tt.wantOK, ok)
			if !tt.wantOK {
				return
			}
			want, err := bson.Marshal(tt.want)
			require.NoError(t, err)
			assert.Equal(t, bson.Raw(want), got)
		})
	}
}
