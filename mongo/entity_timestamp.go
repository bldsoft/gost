package mongo

import "time"

const (
	BsonFieldNameCreateTime   = "createTime"
	BsonFieldNameCreateUserID = "createUserId"
	BsonFieldNameUpdateTime   = "updateTime"
	BsonFieldNameUpdateUserID = "updateUserId"
)

type EntityTimeStamp struct {
	CreateTime   *time.Time  `json:"createTime,omitempty" bson:"createTime,omitempty"`
	CreateUserID interface{} `json:"createUserId,omitempty" bson:"createUserId,omitempty"`
	UpdateTime   *time.Time  `json:"updateTime,omitempty" bson:"updateTime,omitempty"`
	UpdateUserID interface{} `json:"updateUserId,omitempty" bson:"updateUserId,omitempty"`
}

func (e *EntityTimeStamp) SetUpdateFields(updateTime time.Time, updateUserID interface{}) {
	e.UpdateTime = &updateTime
	e.UpdateUserID = updateUserID
}

func (e *EntityTimeStamp) SetCreateFields(createTime time.Time, createUserID interface{}) {
	e.CreateTime = &createTime
	e.CreateUserID = createUserID
}
