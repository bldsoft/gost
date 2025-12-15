package webhook

import "github.com/bldsoft/gost/alert/notify/channel"

type Receiver struct {
	URL string
}

func (r Receiver) IsReceiver() {}

var _ channel.Receiver = Receiver{}
