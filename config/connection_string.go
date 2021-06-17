package config

import (
	"encoding/json"
	"net/url"
)

const passReplacement = "***"

func getQueryPasswordNames() []string {
	return []string{"password", "pass"}
}

type ConnectionString string

func (c *ConnectionString) MarshalJSON() ([]byte, error) {
	str := c.String()
	u, err := url.Parse(str)
	if err != nil {
		return nil, err
	}

	query := u.Query()
	for _, queryName := range getQueryPasswordNames() {
		if password := query.Get(queryName); password != "" {
			query.Set(queryName, passReplacement)
			u.RawQuery = query.Encode()
			break
		}
	}

	if _, exsist := u.User.Password(); exsist {
		u.User = url.UserPassword(u.User.Username(), passReplacement)
	}

	result, err := url.QueryUnescape(u.String())
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

func (c ConnectionString) String() string {
	return string(c)
}

type HidenString string

func (c *HidenString) MarshalJSON() ([]byte, error) {
	return json.Marshal(passReplacement)
}

func (c HidenString) String() string {
	return string(c)
}
