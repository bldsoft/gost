package email

import (
	"context"
	"net/smtp"
	"net/textproto"
	"strings"

	"github.com/bldsoft/gost/alert/notify/channel"
	"github.com/jordan-wright/email"
)

type Email struct {
	Cfg  Config
	auth smtp.Auth
}

func NewEmail(cfg Config) *Email {
	cfg = prepareEmailConfig(cfg)
	return &Email{
		Cfg:  cfg,
		auth: smtp.PlainAuth("", cfg.SMTP.AuthUsername, cfg.SMTP.AuthPassword.String(), cfg.SMTP.Address.Host()),
	}
}

func (e *Email) Send(ctx context.Context, receiver Receiver, message channel.Message) error {
	var subject, body strings.Builder
	err := e.Cfg.MessageTemplate.Execute(&body, message.Data)
	if err != nil {
		return err
	}
	err = e.Cfg.SubjectTemplate.Execute(&subject, message.Data)
	if err != nil {
		return err
	}

	msg := &email.Email{
		To:      receiver.Emails,
		From:    e.Cfg.SMTP.Sender,
		Subject: subject.String(),
		Headers: textproto.MIMEHeader{},
	}
	if e.Cfg.UsePlainText {
		msg.Text = []byte(body.String())
	} else {
		msg.HTML = []byte(body.String())
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return msg.Send(e.Cfg.SMTP.Address.HostPort(), e.auth)
	}
}
