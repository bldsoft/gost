package config

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		{
			"clickhouse://user:qwerty@h1:9000,h2:9000/db",
			"clickhouse://user:***@h1:9000,h2:9000/db",
		},
		{
			"mongodb://user:qwerty@h1:27017,h2:27017/db?replicaSet=rs0&authSource=admin",
			"mongodb://user:***@h1:27017,h2:27017/db?replicaSet=rs0&authSource=admin",
		},
		{
			"clickhouse://h1:9000,h2:9000/db?password=qwerty&username=user",
			"clickhouse://h1:9000,h2:9000/db?password=***&username=user",
		},
		{
			"mongodb://u:p@ss:w@rd@h/db",
			"mongodb://u:***@h/db",
		},
		{
			"clickhouse://user@h:9000/db",
			"clickhouse://user@h:9000/db",
		},
		{
			"clickhouse://h:9000/db?Password=qwerty",
			"clickhouse://h:9000/db?Password=***",
		},
		{
			"clickhouse://h:9000/db?compress=1&password",
			"clickhouse://h:9000/db?compress=1&password",
		},
		{
			"not a url",
			"not a url",
		},
	}
	assert.Equal(t, "", ConnectionString("").Masked())

	for _, test := range testCases {
		t.Run(test.url.String(), func(t *testing.T) {
			assert.Equal(t, test.exp, test.url.Masked())

			b, err := yaml.Marshal(&test.url)
			require.NoError(t, err)
			assert.Equal(t, test.exp, strings.Trim(string(b), "\n"))
		})
	}
}

func TestConnectionStringNeverLeaksPassword(t *testing.T) {
	const secret = "s3cr3t"
	for _, dsn := range []ConnectionString{
		"clickhouse://user:" + secret + "@h1:9000,h2:9000/db",
		"clickhouse://h1:9000,h2:9000/db?password=" + secret,
		"clickhouse://user:" + secret + "@h1:9000,h2:9000/db?pass=" + secret,
		"mongodb://user:" + secret + "@h1,h2,h3/db?replicaSet=rs0",
		"://user:" + secret + "@",
	} {
		b, err := json.Marshal(&dsn)
		require.NoError(t, err)
		assert.NotContains(t, string(b), secret, "%s", dsn)
		assert.NotContains(t, dsn.Masked(), secret, "%s", dsn)
	}
}
