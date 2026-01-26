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
	ColorTemplate   *template.Template // optional attachment color
}

var DefaultWebhookConfig = WebhookConfig{
	MessageTemplate: DefaultMessageTemplate,
}

func prepareWebhookConfig(cfg WebhookConfig) webhook.Config {
	cfg.MessageTemplate = cmp.Or(cfg.MessageTemplate, DefaultWebhookConfig.MessageTemplate)
	return webhook.Config{
		BodyFormat: bodyFormatFunc(cfg.MessageTemplate, cfg.ColorTemplate),
	}
}

func bodyFormatFunc(msgTemplate, colorTemplate *template.Template) func(msg channel.Message) (body []byte, mimeType string) {
	return func(msg channel.Message) (body []byte, mimeType string) {
		var msgTxt strings.Builder
		if err := msgTemplate.Execute(&msgTxt, msg.Data); err != nil {
			return nil, "application/json"
		}

		if colorTemplate != nil {
			var colorTxt strings.Builder
			if err := colorTemplate.Execute(&colorTxt, msg.Data); err == nil && colorTxt.Len() > 0 {
				body, _ = json.Marshal(struct {
					Attachments []struct {
						Color    string   `json:"color"`
						Text     string   `json:"text"`
						MrkdwnIn []string `json:"mrkdwn_in"`
					} `json:"attachments"`
				}{
					Attachments: []struct {
						Color    string   `json:"color"`
						Text     string   `json:"text"`
						MrkdwnIn []string `json:"mrkdwn_in"`
					}{
						{
							Color:    strings.TrimSpace(colorTxt.String()),
							Text:     msgTxt.String(),
							MrkdwnIn: []string{"text"},
						},
					},
				})
				return body, "application/json"
			}
		}

		body, _ = json.Marshal(struct {
			Text string `json:"text"`
		}{
			Text: msgTxt.String(),
		})
		return body, "application/json"
	}
}
