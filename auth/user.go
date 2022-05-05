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
	UserLogin      string `json:"name,omitempty" bson:"name,omitempty"`
	EntityPassword `bson:",inline" json:",inline"`
}

func (c *Creds) Login() string {
	return c.UserLogin
}

type User struct {
	Creds `bson:",inline" json:",inline"`
}

var _ Authenticable = (*User)(nil)
