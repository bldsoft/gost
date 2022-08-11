package changelog

type ILoggedEntity interface {
	SetChangeID(id interface{})
}

type LoggedEntity struct {
	ChangeRecordID idType `json:"changeID,omitempty" bson:"changeID"`
}

func (entity *LoggedEntity) SetChangeID(id idType) {
	entity.ChangeRecordID = id
}
