package auth

const BsonFieldNameEmail = "name"
const BsonFieldNamePassword = "password"
const BsonFieldNameRole = "role"

type EntityPassword struct {
	UserPassword string `json:"password,omitempty" bson:"password,omitempty"`
}

func (c *EntityPassword) SetPassword(password string) {
	c.UserPassword = password
}
func (c *EntityPassword) Password() string {
	return c.UserPassword
}

type Creds struct {
	UserLogin    string `json:"name,omitempty" bson:"name,omitempty"`
	EntityPassword `bson:",inline" json:",inline"`
}

func (c *Creds) Login() string {
	return c.UserLogin
}


type EntityRole[T IRole] struct {
	UserRole T `json:"role,omitempty" bson:"role,omitempty"`
}

func (e *EntityRole[T]) Role() T {
	return e.UserRole
}

type User[R IRole] struct {
	Creds         `bson:",inline" json:",inline"`
	EntityRole[R] `bson:",inline" json:",inline"`
}

var _ Authenticatable = (*User[int])(nil)
var _ Authorizable[int] = (*User[int])(nil)
