package notify

import (
	"context"
	"errors"
	"fmt"

	"github.com/bldsoft/gost/alert/notify/channel"
	"github.com/bldsoft/gost/alert/notify/channel/email"
	"github.com/bldsoft/gost/alert/notify/channel/slack"
	"github.com/bldsoft/gost/alert/notify/channel/webhook"
	"github.com/bldsoft/gost/utils/poly"
)

var ErrUnknownNotificationType = fmt.Errorf("notification type: %w", errors.ErrUnsupported)
var ErrChannelNotConfigured = fmt.Errorf("channel not configured")

type Receiver = channel.Receiver
type Message = channel.Message

func init() {
	poly.Register[Receiver]().
		Type("email", email.Receiver{}).
		Type("slack_webhook", slack.Receiver{}).
		Type("webhook", webhook.Receiver{})
}

type DispatcherConfig struct {
	Email        *email.Config
	SlackWebhook *slack.WebhookConfig
	Webhook      *webhook.Config
}

type Dispatcher struct {
	email   channelWrapper[email.Receiver]
	slack   channelWrapper[slack.Receiver]
	webhook channelWrapper[webhook.Receiver]
}

func NewDispatcher(cfg DispatcherConfig) *Dispatcher {
	d := &Dispatcher{}
	if cfg.Email != nil {
		d.email = channelWrapper[email.Receiver]{
			channel: email.NewEmail(*cfg.Email),
		}
	}
	if cfg.SlackWebhook != nil {
		d.slack = channelWrapper[slack.Receiver]{
			channel: slack.NewWebhook(*cfg.SlackWebhook),
		}
	}
	if cfg.Webhook != nil {
		d.webhook = channelWrapper[webhook.Receiver]{
			channel: webhook.NewWebhook(*cfg.Webhook),
		}
	}
	return d
}

func (d *Dispatcher) Send(ctx context.Context, notification Notification) error {
	switch rcv := notification.Receiver.Value.(type) {
	case email.Receiver:
		return d.email.Send(ctx, rcv, notification.Message)
	case slack.Receiver:
		return d.slack.Send(ctx, rcv, notification.Message)
	case webhook.Receiver:
		return d.webhook.Send(ctx, rcv, notification.Message)
	default:
		return fmt.Errorf("%w: %s", ErrUnknownNotificationType, rcv)
	}
}

type channelWrapper[R Receiver] struct {
	channel channel.Channel[R]
}

func (w *channelWrapper[R]) Send(ctx context.Context, receiver R, message Message) error {
	if w.channel == nil {
		return fmt.Errorf("%w: %T", ErrChannelNotConfigured, receiver)
	}
	return w.channel.Send(ctx, receiver, message)
}
