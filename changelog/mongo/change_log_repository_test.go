//go:build integration_test

package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/bldsoft/gost/changelog"
	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
	"github.com/bldsoft/gost/repository"
)

const (
	bsonFieldNameRequestID   = "requestID"
	sortTestUserCollection   = "spl1428_test_user"
	sortTestRequestIDPattern = "^spl1428-test-"

	sortTestIDZelda   = "spl1428-test-id-aaa"
	sortTestIDMatthew = "spl1428-test-id-mmm"
	sortTestIDAaron   = "spl1428-test-id-zzz"
)

func connectTestStorage(t *testing.T) *mongo.Storage {
	t.Helper()
	log.SetLogLevel("")

	var cfg mongo.Config
	config.ReadConfig(&cfg, "")

	db := mongo.NewStorage(cfg)
	db.Connect()
	t.Cleanup(func() { _ = db.Disconnect(context.Background()) })

	return db
}

func seedSortByUserName(t *testing.T, ctx context.Context, db *mongo.Storage, rep *ChangeLogRepository) int {
	t.Helper()

	cleanup := func() {
		_, err := db.Db.Collection("change_log").DeleteMany(ctx, bson.M{
			bsonFieldNameRequestID: bson.M{"$regex": sortTestRequestIDPattern},
		})
		assert.NoError(t, err)
		assert.NoError(t, db.Db.Collection(sortTestUserCollection).Drop(ctx))
	}
	cleanup()
	t.Cleanup(cleanup)

	_, err := db.Db.Collection(sortTestUserCollection).InsertMany(ctx, []any{
		bson.M{"_id": sortTestIDZelda, "name": "Zelda"},
		bson.M{"_id": sortTestIDMatthew, "name": "Matthew"},
		bson.M{"_id": sortTestIDAaron, "name": "aaron"},
	})
	require.NoError(t, err)

	now := int(time.Now().Unix())
	newRecord := func(userID, name string, timestamp int) *Record {
		rec := &Record{Record: &changelog.Record{
			UserID:    userID,
			Timestamp: int64(timestamp),
			Operation: changelog.Create,
			Entity:    "session",
			EntityID:  userID,
			RequestID: "spl1428-test-" + name,
		}}
		rec.GenerateID()

		return rec
	}
	require.NoError(t, rep.InsertMany(ctx, []*Record{
		newRecord(sortTestIDAaron, "aaron", now-3000),
		newRecord(sortTestIDZelda, "Zelda", now-2000),
		newRecord(sortTestIDMatthew, "Matthew", now-1000),
	}))

	return now
}

func TestChangeLogRepository_SortByUserName(t *testing.T) {
	ctx := context.Background()
	db := connectTestStorage(t)

	rep := NewChangeLogRepository(db)
	rep.SetNameSortJoin(NameSortJoin{
		From:         sortTestUserCollection,
		LocalField:   changelog.BsonFieldNameUserID,
		ForeignField: "_id",
		NameField:    "name",
	})
	now := seedSortByUserName(t, ctx, db, rep)
	repWithoutJoin := NewChangeLogRepository(db)

	testRequestIDs := []string{"spl1428-test-aaron", "spl1428-test-Zelda", "spl1428-test-Matthew"}
	onlyTestRows := func(f changelog.Filter) *changelog.Filter {
		f.Details = requestIDFilter(testRequestIDs)
		return &f
	}
	from, to := now-1500, now-500

	tests := []struct {
		name           string
		rep            *ChangeLogRepository
		params         changelog.RecordsParams
		wantUserIDs    []string
		wantTotalCount int64
	}{
		{
			name:           "user ASC is case-insensitive name order",
			rep:            rep,
			params:         changelog.RecordsParams{Filter: onlyTestRows(changelog.Filter{}), Sort: changelog.Sort{Field: changelog.SortFieldUser, Order: repository.SortOrderASC}},
			wantUserIDs:    []string{sortTestIDAaron, sortTestIDMatthew, sortTestIDZelda},
			wantTotalCount: 3,
		},
		{
			name:           "user DESC is reverse name order",
			rep:            rep,
			params:         changelog.RecordsParams{Filter: onlyTestRows(changelog.Filter{}), Sort: changelog.Sort{Field: changelog.SortFieldUser, Order: repository.SortOrderDESC}},
			wantUserIDs:    []string{sortTestIDZelda, sortTestIDMatthew, sortTestIDAaron},
			wantTotalCount: 3,
		},
		{
			name:           "first page keeps total count",
			rep:            rep,
			params:         changelog.RecordsParams{Filter: onlyTestRows(changelog.Filter{}), Sort: changelog.Sort{Field: changelog.SortFieldUser, Order: repository.SortOrderASC}, Limit: 2},
			wantUserIDs:    []string{sortTestIDAaron, sortTestIDMatthew},
			wantTotalCount: 3,
		},
		{
			name:           "second page continues name order",
			rep:            rep,
			params:         changelog.RecordsParams{Filter: onlyTestRows(changelog.Filter{}), Sort: changelog.Sort{Field: changelog.SortFieldUser, Order: repository.SortOrderASC}, Limit: 2, Offset: 2},
			wantUserIDs:    []string{sortTestIDZelda},
			wantTotalCount: 3,
		},
		{
			name:           "user IDs filter is applied",
			rep:            rep,
			params:         changelog.RecordsParams{Filter: onlyTestRows(changelog.Filter{UserIDs: []string{sortTestIDZelda}}), Sort: changelog.Sort{Field: changelog.SortFieldUser, Order: repository.SortOrderASC}},
			wantUserIDs:    []string{sortTestIDZelda},
			wantTotalCount: 1,
		},
		{
			name:           "time range filter is applied",
			rep:            rep,
			params:         changelog.RecordsParams{Filter: onlyTestRows(changelog.Filter{From: &from, To: &to}), Sort: changelog.Sort{Field: changelog.SortFieldUser, Order: repository.SortOrderASC}},
			wantUserIDs:    []string{sortTestIDMatthew},
			wantTotalCount: 1,
		},
		{
			name:           "timestamp sort is unchanged with join",
			rep:            rep,
			params:         changelog.RecordsParams{Filter: onlyTestRows(changelog.Filter{}), Sort: changelog.Sort{Field: changelog.SortFieldTimestamp, Order: repository.SortOrderDESC}},
			wantUserIDs:    []string{sortTestIDMatthew, sortTestIDZelda, sortTestIDAaron},
			wantTotalCount: 3,
		},
		{
			name:           "user sort without join stays user ID order",
			rep:            repWithoutJoin,
			params:         changelog.RecordsParams{Filter: onlyTestRows(changelog.Filter{}), Sort: changelog.Sort{Field: changelog.SortFieldUser, Order: repository.SortOrderASC}},
			wantUserIDs:    []string{sortTestIDZelda, sortTestIDMatthew, sortTestIDAaron},
			wantTotalCount: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.rep.GetRecords(ctx, &tt.params)
			require.NoError(t, err)
			require.NotNil(t, res)

			userIDs := make([]string, 0, len(res.Records))
			for _, rec := range res.Records {
				userIDs = append(userIDs, rec.UserID)
			}
			assert.Equal(t, tt.wantUserIDs, userIDs)
			assert.Equal(t, tt.wantTotalCount, res.TotalCount)
		})
	}
}

type requestIDFilter []string

func (f requestIDFilter) Filter(filter any) {
	filter.(bson.M)[bsonFieldNameRequestID] = bson.M{"$in": []string(f)}
}
