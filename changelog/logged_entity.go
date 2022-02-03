package changelog

type ILoggedEntity interface {
	SetChangeID(id interface{})
}

type LoggedEntity struct {
	ChangeRecordID idType `json:"changeID" bson:"changeID"`
}

func (entity *LoggedEntity) SetChangeID(id idType) {
	entity.ChangeRecordID = id
}
