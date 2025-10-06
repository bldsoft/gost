package webhook

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/bldsoft/gost/alert/notify/channel"
)

type Webhook struct {
	Cfg Config
}

func NewWebhook(cfg Config) *Webhook {
	cfg = prepareWebhookConfig(cfg)
	return &Webhook{
		Cfg: cfg,
	}
}

func (w *Webhook) client() *http.Client {
	return http.DefaultClient
}

func (w *Webhook) Send(ctx context.Context, receiver Receiver, msg channel.Message) error {
	body, mimeType := w.Cfg.BodyFormat(msg)
	resp, err := w.client().Post(
		receiver.URL,
		mimeType,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("%d: %s", resp.StatusCode, string(body))
	}
	return nil
}
