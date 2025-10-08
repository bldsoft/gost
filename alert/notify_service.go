package alert

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"text/template"

	"github.com/bldsoft/gost/alert/notify"
	"github.com/bldsoft/gost/alert/notify/channel"
	"github.com/bldsoft/gost/utils/poly"
	"github.com/bldsoft/gost/utils/seq"

	_ "embed"
)

//go:embed notify_templates/email_subject.tmpl
var emailSubjectTemplate string

//go:embed notify_templates/email_message.tmpl
var emailMessageTemplate string

//go:embed notify_templates/slack.tmpl
var slackMessageTemplate string

var defaultEmailSubjectTemplate = template.Must(template.New("email_subject").Parse(emailSubjectTemplate))
var defaultEmailMessageTemplate = template.Must(template.New("email_message").Parse(emailMessageTemplate))
var defaultSlackMessageTemplate = template.Must(template.New("slack_message").Parse(slackMessageTemplate))

const (
	FromMsgKey     = "from"
	ToMsgKey       = "to"
	SeverityMsgKey = "severity"
)

type NotifyConfig = notify.ServiceConfig

type NotifyServiceAdapter struct {
	cfg           NotifyConfig
	notifyService *notify.Service
	receivers     []poly.Poly[notify.Receiver]
}

func NewNotifyService(cfg NotifyConfig, receivers ...poly.Poly[notify.Receiver]) *NotifyServiceAdapter {
	cfg.Dispatcher.Email.MessageTemplate = cmp.Or(cfg.Dispatcher.Email.MessageTemplate, defaultEmailMessageTemplate)
	cfg.Dispatcher.Email.SubjectTemplate = cmp.Or(cfg.Dispatcher.Email.SubjectTemplate, defaultEmailSubjectTemplate)
	cfg.Dispatcher.SlackWebhook.MessageTemplate = cmp.Or(cfg.Dispatcher.SlackWebhook.MessageTemplate, defaultSlackMessageTemplate)
	return &NotifyServiceAdapter{
		cfg:           cfg,
		notifyService: notify.NewService(cfg),
		receivers:     receivers,
	}
}

func (s *NotifyServiceAdapter) Run(ctx context.Context) error {
	return s.notifyService.Run(ctx)
}

func (s *NotifyServiceAdapter) SetQueue(queue notify.Queue) *NotifyServiceAdapter {
	_ = s.notifyService.SetQueue(queue)
	return s
}

func (s *NotifyServiceAdapter) Send(ctx context.Context, alert Alert) error {
	var errs error
	for _, receiver := range seq.Concat2(
		slices.All(s.receivers),
		slices.All(alert.Receivers),
	) {
		notification := notify.Notification{
			Receiver: receiver,
			Message:  s.prepareMessage(alert),
		}
		err := s.notifyService.Send(ctx, notification)
		errs = errors.Join(errs, err)
	}
	return errs
}

func (h *NotifyServiceAdapter) prepareMessage(alert Alert) channel.Message {
	msg := channel.Message{
		Data: make(map[string]any),
	}
	for k, v := range alert.MetaData {
		msg.Data[k] = v
	}
	msg.Data[FromMsgKey] = alert.From
	msg.Data[ToMsgKey] = alert.To
	msg.Data[SeverityMsgKey] = alert.Severity
	return msg
}
