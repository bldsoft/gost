package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Param1 string
	Param2 int
	Param3 bool

	TaggedField string `mapstructure:"TAGGED"`

	Mongo      Database   `mapstructure:"MONGO"`
	Clickhouse Clickhouse `mapstructure:"CLICKHOUSE"`
}

func (c *Config) SetDefaults() {
	c.Clickhouse.Database.ConnectionString = "tcp://127.0.0.1:9000?database=Parent"
	c.Clickhouse.Param1 = "parent"
}

func (c *Config) Validate() error { return nil }

type Clickhouse struct {
	Database
	Param1 string `mapstructure:"PARAM1"`
	Param2 string `mapstructure:"PARAM2"`
}

func (c *Clickhouse) SetDefaults() {
	c.Database.ConnectionString = "tcp://127.0.0.1:9000?database=Nested"
	c.Param1 = "nested"
	c.Param2 = "nested"
}

func (c *Clickhouse) Validate() error { return nil }

type Database struct {
	ConnectionString string `mapstructure:"CONNECTION_STRING"`
}

func Example() {
	var cfg Config
	defer os.Clearenv()
	os.Setenv("EX_PARAM1", "gost")
	os.Setenv("EX_PARAM2", "2")
	os.Setenv("EX_PARAM3", "true")

	os.Setenv("EX_TAGGED", "tagged value")

	os.Setenv("EX_MONGO_CONNECTION_STRING", "mongodb://localhost:27017")
	os.Setenv("EX_CLICKHOUSE_CONNECTION_STRING", "tcp://127.0.0.1:9000?database=test")
	// The value of EX_CLICKHOUSE_PARAM1 will be set from Config.SetDefaults (not Clickhouse.SetDefaults)
	// The value of EX_CLICKHOUSE_PARAM2 will be set from Clickhouse.SetDefaults

	ReadConfig(&cfg, "EX")

	data, _ := json.MarshalIndent(cfg, "", "	")
	fmt.Print(string(data))

	// Output:
	//{
	//	"Param1": "gost",
	//	"Param2": 2,
	//	"Param3": true,
	//	"TaggedField": "tagged value",
	//	"Mongo": {
	//		"ConnectionString": "mongodb://localhost:27017"
	//	},
	//	"Clickhouse": {
	//		"ConnectionString": "tcp://127.0.0.1:9000?database=test",
	//		"Param1": "parent",
	//		"Param2": "nested"
	//	}
	//}
}
