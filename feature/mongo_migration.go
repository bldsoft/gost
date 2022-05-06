package feature

import (
	"encoding/json"
	"fmt"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
)

func AddFeatureMigration(db *mongo.Storage, version uint, features ...*Feature) {
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
	"insert": "feature",
	"documents": %s
	},
	{
	"update": "feature",
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
]`, featureStr, IDs, mongo.BsonFieldNameCreateTime, mongo.BsonFieldNameUpdateTime)

	down := fmt.Sprintf(`[{
	"delete": "feature",
	"deletes": [{
		"q": {
			"_id": { "$in" : [%s] }
		},
		"limit": 0
	}]
}]`, IDs)

	db.AddMigration(version, up, down)
}
