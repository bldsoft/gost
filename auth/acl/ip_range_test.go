package acl

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

type acl struct {
	ACL IpRange `json:"acl" bson:"acl"`
}

func TestIpRangeJson(t *testing.T) {
	var acl acl
	t.Run("Unmarshal", func(t *testing.T) {
		assert.NoError(t, json.Unmarshal([]byte(`{ "acl": ["127.0.0.0/24","192.168.0.1"]}`), &acl))
		assert.True(t, acl.ACL.Contains(net.ParseIP("127.0.0.1")))
		assert.True(t, acl.ACL.Contains(net.ParseIP("192.168.0.1")))
		assert.False(t, acl.ACL.Contains(net.ParseIP("192.168.0.2")))
	})
	t.Run("Marshal", func(t *testing.T) {
		data, err := json.Marshal(acl)
		assert.NoError(t, err)
		assert.Equal(t, `{"acl":["192.168.0.1","127.0.0.0/24"]}`, string(data))
	})
}

func TestIpRangeBson(t *testing.T) {
	acl := acl{
		ACL: MustIpRangeFromStrings("127.0.0.0/24", "192.168.0.1"),
	}

	data, err := bson.Marshal(acl)
	assert.NoError(t, err)

	assert.NoError(t, bson.Unmarshal(data, &acl))
	assert.True(t, acl.ACL.Contains(net.ParseIP("127.0.0.1")))
	assert.True(t, acl.ACL.Contains(net.ParseIP("192.168.0.1")))
	assert.False(t, acl.ACL.Contains(net.ParseIP("192.168.0.2")))
}
