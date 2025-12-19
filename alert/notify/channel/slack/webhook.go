package slack

import (
	"context"

	"github.com/bldsoft/gost/alert/notify/channel"
	"github.com/bldsoft/gost/alert/notify/channel/webhook"
)

type Webhook struct {
	webhook.Webhook
}

func NewWebhook(cfg WebhookConfig) *Webhook {
	webhookConfig := prepareWebhookConfig(cfg)
	return &Webhook{
		Webhook: webhook.Webhook{Cfg: webhookConfig},
	}
}

func (w *Webhook) Send(ctx context.Context, receiver Receiver, msg channel.Message) error {
	return w.Webhook.Send(ctx, webhook.Receiver{URL: receiver.URL}, msg)
}
