//go:build integration_test

package alert

import (
	"context"
	"testing"
	"time"

	"github.com/bldsoft/gost/alert/notify"
	"github.com/bldsoft/gost/alert/notify/channel/email"
	"github.com/bldsoft/gost/alert/notify/channel/slack"
	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/utils/poly"
)

type testConfig struct {
	SMTP            email.SMTPConfig `mapstructure:"SMTP"`
	SlackWebhookURL string           `mapstructure:"SLACK_WEBHOOK_URL"`
	EmailReceivers  []string         `mapstructure:"EMAIL_RECEIVERS"`
}

func (c *testConfig) SetDefaults()    {}
func (c *testConfig) Validate() error { return nil }

var cfg testConfig

func init() {
	config.ReadConfig(&cfg, "")
}

func TestNotifyHandler(t *testing.T) {
	service := NewNotifyService(NotifyConfig{
		RetryCount:             1,
		RetryQueuePollInterval: 1 * time.Second,
		WorkerN:                1,
		Dispatcher: notify.DispatcherConfig{
			Email: &email.Config{
				SMTP: cfg.SMTP,
			},
			SlackWebhook: &slack.WebhookConfig{},
		},
	})
	handler := NotifyHandler(service)

	handler.Handle(context.Background(), Alert{
		SourceID: "test",
		Severity: SeverityHigh,
		From:     time.Now(),
		To:       time.Now().Add(1 * time.Hour),
		MetaData: map[string]string{
			"description": "description",
			"test":        "test",
		},
		Receivers: []poly.Poly[notify.Receiver]{
			{Value: slack.Receiver{
				URL: cfg.SlackWebhookURL,
			}},
			{Value: email.Receiver{
				Emails: cfg.EmailReceivers,
			}},
		},
	})
}
