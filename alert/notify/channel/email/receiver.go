package email

import "github.com/bldsoft/gost/alert/notify/channel"

type Receiver struct {
	Emails []string
}

func (r Receiver) IsReceiver() {}

var _ channel.Receiver = Receiver{}
