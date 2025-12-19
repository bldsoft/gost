package email

import (
	"cmp"
	"text/template"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/utils"

	_ "embed"
)

//go:embed default_message.tmpl
var emailTemplate string
var DefaultEmailTemplate = template.Must(template.New("email").Parse(emailTemplate))

//go:embed default_subject.tmpl
var subjectTemplate string
var DefaultSubjectTemplate = template.Must(template.New("subject").Parse(subjectTemplate))

type Config struct {
	SMTP            SMTPConfig
	SubjectTemplate *template.Template
	MessageTemplate *template.Template
	UsePlainText    bool
}

type SMTPConfig struct {
	Address      config.Address    `mapstructure:"ADDRESS" description:"SMTP server address"`
	AuthUsername string            `mapstructure:"USERNAME" description:"SMTP admin username"`
	AuthPassword utils.FullyHidden `mapstructure:"PASSWORD" description:"SMTP admin password"`
	Sender       string            `mapstructure:"SENDER" description:"SMTP email sender"`
}

var DefaultEmailConfig = Config{
	SMTP: SMTPConfig{
		Address:      config.Address("smtp.gmail.com:587"),
		AuthUsername: "",
		AuthPassword: "",
		Sender:       "",
	},
	SubjectTemplate: DefaultSubjectTemplate,
	MessageTemplate: DefaultEmailTemplate,
	UsePlainText:    false,
}

func prepareEmailConfig(cfg Config) Config {
	cfg.SMTP.Address = cmp.Or(cfg.SMTP.Address, DefaultEmailConfig.SMTP.Address)
	cfg.SMTP.AuthUsername = cmp.Or(cfg.SMTP.AuthUsername, DefaultEmailConfig.SMTP.AuthUsername)
	cfg.SMTP.AuthPassword = cmp.Or(cfg.SMTP.AuthPassword, DefaultEmailConfig.SMTP.AuthPassword)
	cfg.SMTP.Sender = cmp.Or(cfg.SMTP.Sender, DefaultEmailConfig.SMTP.Sender)
	cfg.MessageTemplate = cmp.Or(cfg.MessageTemplate, DefaultEmailConfig.MessageTemplate)
	cfg.SubjectTemplate = cmp.Or(cfg.SubjectTemplate, DefaultEmailConfig.SubjectTemplate)
	return cfg
}
