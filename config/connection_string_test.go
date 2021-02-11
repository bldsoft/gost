package config

import (
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

func TestConnectionString(t *testing.T) {
	testCases := []struct {
		url ConnectionString
		exp string
	}{
		{
			"mongodb://user:qwerty@127.0.0.1:27017/streampool",
			"mongodb://user:***@127.0.0.1:27017/streampool",
		},
		{
			"tcp://us-click.dev.spnode.net:9000?database=streampool&password=qwerty&username=user",
			"tcp://us-click.dev.spnode.net:9000?database=streampool&password=***&username=user",
		},
		{
			"tcp://us-click.dev.spnode.net:9000?database=streampool&pass=qwerty&username=user",
			"tcp://us-click.dev.spnode.net:9000?database=streampool&pass=***&username=user",
		},
		{
			"tcp://us-click.dev.spnode.net:9000?database=streampool&password=password&username=user",
			"tcp://us-click.dev.spnode.net:9000?database=streampool&password=***&username=user",
		},
		{
			"mongodb://pass:pass@127.0.0.1:27017/streampool",
			"mongodb://pass:***@127.0.0.1:27017/streampool",
		},
		{
			"mongodb://mongo:mongo@127.0.0.1:27017/mongo",
			"mongodb://mongo:***@127.0.0.1:27017/mongo",
		},
		{
			"scheme://user:pass@127.0.0.1:27017/path?pass=pass",
			"scheme://user:***@127.0.0.1:27017/path?pass=***",
		},
	}

	for _, test := range testCases {
		t.Run(test.url.String(), func(t *testing.T) {
			b, _ := yaml.Marshal(&test.url)
			u := strings.Trim(string(b), "\n")
			assert.Equal(t, test.exp, u)
		})
	}
}
