package slack

import (
	"cmp"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/bldsoft/gost/alert/notify/channel"
	"github.com/bldsoft/gost/alert/notify/channel/webhook"

	_ "embed"
)

//go:embed default_message.tmpl
var messageTemplate string
var DefaultMessageTemplate = template.Must(template.New("message").Parse(messageTemplate))

type WebhookConfig struct {
	MessageTemplate *template.Template
}

var DefaultWebhookConfig = WebhookConfig{
	MessageTemplate: DefaultMessageTemplate,
}

func prepareWebhookConfig(cfg WebhookConfig) webhook.Config {
	cfg.MessageTemplate = cmp.Or(cfg.MessageTemplate, DefaultWebhookConfig.MessageTemplate)
	return webhook.Config{
		BodyFormat: bodyFormatFunc(cfg.MessageTemplate),
	}
}

func bodyFormatFunc(t *template.Template) func(msg channel.Message) (body []byte, mimeType string) {
	return func(msg channel.Message) (body []byte, mimeType string) {
		var msgTxt strings.Builder
		err := t.Execute(&msgTxt, msg.Data)
		if err != nil {
			return nil, "application/json"
		}
		body, _ = json.Marshal(struct {
			Text string `json:"text"`
		}{
			Text: msgTxt.String(),
		})
		return body, "application/json"
	}
}
