package entity

import "github.com/bldsoft/gost/mongo"

type Note struct {
	mongo.EntityID        `bson:",inline"`
	mongo.EntityTimeStamp `bson:",inline"`
	mongo.EntityArchived  `bson:",inline"`

	Title string `json:"title" bson:"title"`
	Body  string `json:"body,omitempty" bson:"body,omitempty"`
}
