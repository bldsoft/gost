package microservice

import (
	"fmt"

	"github.com/bldsoft/gost/examples/notes/pkg/repository"
	"github.com/bldsoft/gost/mongo"
)

const (
	notesTitleIndexUp = `[{
	"createIndexes": "%s",
	"indexes": [{
		"key": {"title": 1},
		"name": "title_asc"
	}]
}]`
	notesTitleIndexDown = `[{
	"dropIndexes": "%s",
	"index": "title_asc"
}]`
)

// Migrations must be registered before Storage.Connect: Connect runs them and never looks at later additions.
func addMigrations(db *mongo.Storage) {
	db.AddMigration(1,
		fmt.Sprintf(notesTitleIndexUp, repository.CollectionName),
		fmt.Sprintf(notesTitleIndexDown, repository.CollectionName))
}
