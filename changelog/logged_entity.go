package changelog

type ILoggedEntity interface {
	SetChangeID(id string)
}

type LoggedEntity struct {
	ChangeRecordID string `json:"changeID,omitempty" bson:"changeID"`
}

func (entity *LoggedEntity) SetChangeID(id string) {
	entity.ChangeRecordID = id
}
