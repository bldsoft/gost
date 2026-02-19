package feature

import (
	"encoding/json"
	"fmt"

	"github.com/bldsoft/gost/config/feature"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
)

type FeatureMigrator struct {
	db       *mongo.Storage
	collName string
}

func NewFeatureMigrator(db *mongo.Storage, collName string) *FeatureMigrator {
	return &FeatureMigrator{db: db, collName: collName}
}

func (m *FeatureMigrator) AddFeatureMigration(version uint, features ...*Feature) {
	size := len(features)
	if size == 0 {
		return
	}

	featureStr, err := json.Marshal(features)
	if err != nil {
		log.Fatalf("Falied to marshal features for migration: %s", err.Error())
	}

	IDs := fmt.Sprintf("%d", features[0].ID)
	for i := 1; i < size; i++ {
		IDs += fmt.Sprintf(",%d", features[i].ID)
	}

	up := fmt.Sprintf(`[{
	"insert": "%s",
	"documents": %s
	},
	{
	"update": "%s",
	"updates": [{
		"q": {
			"_id": { "$in" : [%s] }
		},
		"u": {
			"$currentDate": {
				"%s": true,
				"%s": true			
			}
		},
		"multi": true
	}]
	}
]`, m.collName, featureStr, m.collName, IDs, mongo.BsonFieldNameCreateTime, mongo.BsonFieldNameUpdateTime)

	down := fmt.Sprintf(`[{
	"delete": "%s",
	"deletes": [{
		"q": {
			"_id": { "$in" : [%s] }
		},
		"limit": 0
	}]
}]`, m.collName, IDs)

	m.db.AddMigration(version, up, down)
}

func (m *FeatureMigrator) DeleteFeatureMigration(version uint, featureIDs ...feature.IdType) {
	if len(featureIDs) == 0 {
		return
	}
	IDs := fmt.Sprintf("%d", featureIDs[0])
	for i := 1; i < len(featureIDs); i++ {
		IDs += fmt.Sprintf(",%d", featureIDs[i])
	}

	m.db.AddMigration(version, fmt.Sprintf(`[{
		"update": "%s",
		"updates": [{
			"q": {
				"_id": { "$in" : [%s] }
			},
			"u": {
				"$set": {
					"%s": true
				}
			},
			"multi": true
		}]
	}]`, m.collName, IDs, mongo.BsonFieldNameArchived), fmt.Sprintf(`[{
		"update": "%s",
		"updates": [{
			"q": {
				"_id": { "$in" : [%s] }
			},
			"u": {
				"$unset": {
					"%s": ""
				}
			},
			"multi": true
		}]
	}]`, m.collName, IDs, mongo.BsonFieldNameArchived))
}

func DeleteFeatureMigration(db *mongo.Storage, version uint, featureIDs ...feature.IdType) {
	m := NewFeatureMigrator(db, DefaultCollectionName)
	m.DeleteFeatureMigration(version, featureIDs...)
}

func AddFeatureMigration(db *mongo.Storage, version uint, features ...*Feature) {
	m := NewFeatureMigrator(db, DefaultCollectionName)
	m.AddFeatureMigration(version, features...)
}
