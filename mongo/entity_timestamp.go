package mongo

import "time"

const (
	BsonFieldNameCreateTime   = "createTime"
	BsonFieldNameCreateUserID = "createUserId"
	BsonFieldNameUpdateTime   = "updateTime"
	BsonFieldNameUpdateUserID = "updateUserId"
)

type EntityTimeStamp struct {
	CreateTime   *time.Time `json:"createTime,omitempty" bson:"createTime,omitempty"`
	CreateUserID any        `json:"createUserId,omitempty" bson:"createUserId,omitempty"`
	UpdateTime   *time.Time `json:"updateTime,omitempty" bson:"updateTime,omitempty"`
	UpdateUserID any        `json:"updateUserId,omitempty" bson:"updateUserId,omitempty"`
}

func (e *EntityTimeStamp) SetUpdateFields(updateTime time.Time, updateUserID any) {
	e.UpdateTime = &updateTime
	if updateUserID != nil {
		e.UpdateUserID = updateUserID
	}
}

func (e *EntityTimeStamp) SetCreateFields(createTime time.Time, createUserID any) {
	e.CreateTime = &createTime
	if createUserID != nil {
		e.CreateUserID = createUserID
	}
}
