package config

import (
	"encoding/json"
	"strings"
)

const passReplacement = "***"

func getQueryPasswordNames() []string {
	return []string{"password", "pass"}
}

// ConnectionString is a DSN whose JSON form hides the password, so it can be logged and shown by /env.
type ConnectionString string

func (c *ConnectionString) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Masked())
}

func (c ConnectionString) String() string {
	return string(c)
}

// Masked returns the DSN with the password of the userinfo and of the password/pass query
// parameters replaced by "***". It never fails: DSNs with comma-separated hosts are not valid
// URLs for net/url, and a parse error would carry the raw string into the log.
func (c ConnectionString) Masked() string {
	s := string(c)
	head, query, hasQuery := strings.Cut(s, "?")
	if hasQuery {
		query = maskQuery(query)
	}

	scheme, rest, hasScheme := strings.Cut(head, "://")
	if hasScheme {
		authority, path, hasPath := strings.Cut(rest, "/")
		if at := strings.LastIndex(authority, "@"); at >= 0 {
			if user, _, hasPass := strings.Cut(authority[:at], ":"); hasPass {
				authority = user + ":" + passReplacement + authority[at:]
			}
		}
		head = scheme + "://" + authority
		if hasPath {
			head += "/" + path
		}
	}
	if hasQuery {
		return head + "?" + query
	}

	return head
}

func maskQuery(query string) string {
	params := strings.Split(query, "&")
	for i, param := range params {
		key, _, hasValue := strings.Cut(param, "=")
		if !hasValue {
			continue
		}
		for _, name := range getQueryPasswordNames() {
			if strings.EqualFold(key, name) {
				params[i] = key + "=" + passReplacement

				break
			}
		}
	}

	return strings.Join(params, "&")
}
