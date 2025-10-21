package webhook

import (
	"encoding/json"

	_ "embed"

	"github.com/bldsoft/gost/alert/notify/channel"
)

type Config struct {
	BodyFormat func(msg channel.Message) (body []byte, mimeType string)
}

var DefaultWebhookConfig = Config{
	BodyFormat: DefaultJSONBodyFormat,
}

func prepareWebhookConfig(cfg Config) Config {
	if cfg.BodyFormat == nil {
		cfg.BodyFormat = DefaultWebhookConfig.BodyFormat
	}
	return cfg
}

type DefaultMessageType struct {
	Message string `json:"message"`
}

func DefaultJSONBodyFormat(msg channel.Message) (body []byte, mimeType string) {
	body, _ = json.Marshal(msg.Data)
	return body, "application/json"
}
