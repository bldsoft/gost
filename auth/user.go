package auth

const (
	BsonFieldNameEmail    = "name"
	BsonFieldNamePassword = "password"
	BsonFieldNameRole     = "role"
)

type EntityPassword struct {
	UserPassword   string `json:"password,omitempty" bson:"password,omitempty"`
	ChangePassword bool   `json:"changePasswordRequired,omitempty" bson:"changePasswordRequired"`
}

func (c *EntityPassword) SetPassword(password string) {
	c.UserPassword = password
}

func (c *EntityPassword) Password() string {
	return c.UserPassword
}

func (c *EntityPassword) ChangePasswordRequired() bool {
	return c.ChangePassword
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
